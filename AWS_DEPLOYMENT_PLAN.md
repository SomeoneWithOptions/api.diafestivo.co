# AWS EC2 Deployment Plan

## Goal

Deploy the application from Amazon ECR to a single EC2 instance running Docker, with Nginx terminating HTTPS and proxying requests to the container.

GitHub Actions will perform testing, image publishing, and deployment. AWS Systems Manager (SSM) Run Command will execute deployment commands on EC2, so no SSH key or public SSH port is required.

## Deployment branch policy

The AWS deployment must happen **only after a push to the `aws` branch**.

- Push to `aws`: test, build, push the commit-tagged image to ECR, then deploy it to EC2.
- Push to `main`: existing CI/image publishing may continue, but it must never deploy to EC2.
- Manual `workflow_dispatch`: it must never deploy to EC2.
- The deployment job must include an explicit guard:

  ```yaml
  if: github.event_name == 'push' && github.ref == 'refs/heads/aws'
  ```

- Deploy the immutable `${GITHUB_SHA}-arm64` image produced from the `aws` branch, never `latest` or the mutable `arm64` tag.
- Configure deployment concurrency so two pushes cannot update the instance simultaneously.
- Use a GitHub environment such as `aws-production` for deployment variables and optional approval rules.

The existing workflow can listen to both `main` and `aws`, while only the deployment job has permission to update EC2. Merging `main` into `aws` and pushing the result becomes the controlled production release mechanism.

## Current repository readiness

The repository already provides:

- A production multi-stage `Dockerfile`
- A non-root runtime container listening on port `3002`
- A `GET /healthz` endpoint
- GitHub Actions tests and AWS authentication through GitHub OIDC
- ECR images tagged as `${GITHUB_SHA}-arm64` and `${GITHUB_SHA}-amd64`

No application code change or additional health endpoint should be needed.

## Architecture

```text
Internet
   |
   v
Elastic IP -> EC2 security group (80/443 only)
   |
   v
Nginx + Let's Encrypt TLS
   |
   v
127.0.0.1:3002
   |
   v
Docker container pulled from Amazon ECR
```

GitHub Actions communicates with the instance through AWS SSM rather than SSH.

## 1. Provision the EC2 infrastructure

Create:

- One Ubuntu 24.04 ARM64 EC2 instance
- A `t4g.micro` or `t4g.small` instance, depending on expected traffic
- A small encrypted EBS volume
- An Elastic IP
- A DNS `A` record for `api.diafestivo.co` pointing to the Elastic IP
- A security group allowing:
  - TCP `80` from the internet
  - TCP `443` from the internet
  - No public TCP `22`
  - No public TCP `3002`

ARM64 is preferred because the current workflow already produces a native ARM64 image. If an AMD64 instance is selected instead, use the `${GITHUB_SHA}-amd64` image and update the deployment dependency accordingly.

## 2. Configure the EC2 IAM role

Attach an instance profile with:

- `AmazonSSMManagedInstanceCore`
- Minimal ECR pull permissions:
  - `ecr:GetAuthorizationToken`
  - `ecr:BatchCheckLayerAvailability`
  - `ecr:BatchGetImage`
  - `ecr:GetDownloadUrlForLayer`

Scope repository-specific ECR permissions to the `api-diafestivo` repository where AWS permits it. No permanent AWS credentials should be stored on the instance.

## 3. Update the GitHub Actions IAM role

The existing GitHub OIDC role already pushes images to ECR. Add only the permissions required to:

- Send an SSM command to the production EC2 instance
- Read the command status and output

Restrict `ssm:SendCommand` to the selected instance and the required AWS SSM document where possible.

Store the EC2 instance ID as a GitHub `aws-production` environment variable, not as a secret. Continue using OIDC rather than AWS access keys.

## 4. Bootstrap the instance once

Install and enable:

- Docker
- Nginx
- AWS CLI
- AWS SSM Agent
- Certbot and its Nginx integration
- `curl` for local health checks

Create a small server-side deployment script that GitHub Actions invokes through SSM.

Store application configuration in a root-readable file:

```text
/etc/api-diafestivo.env
```

It may contain:

```text
IP_INFO_TOKEN=...
GIPHY_KEY=...
MY_CIDR=...
```

Set restrictive permissions on this file and do not pass these values through GitHub Actions or SSM command arguments.

## 5. Configure Nginx

Nginx should:

- Redirect HTTP to HTTPS
- Terminate TLS
- Proxy to `http://127.0.0.1:3002`
- Forward `Host`
- Set `X-Forwarded-Proto`
- Set `X-Forwarded-For` to `$remote_addr`

The application trusts the first `X-Forwarded-For` value, so Nginx should overwrite client-provided values instead of appending to them.

Run the container with a local-only port binding:

```text
127.0.0.1:3002:3002
```

This ensures the application can only be reached through Nginx.

## 6. Configure HTTPS

After DNS points to the Elastic IP:

1. Verify the HTTP Nginx virtual host is reachable.
2. Use Certbot to issue a certificate for `api.diafestivo.co`.
3. Enable the HTTP-to-HTTPS redirect.
4. Verify the Certbot systemd renewal timer.

## 7. Extend the GitHub Actions workflow

Update `.github/workflows/push-to-ecr.yaml` so pushes to both `main` and `aws` can run the existing test/build jobs.

Add a `deploy-aws` job that:

1. Depends on the ARM64 build-and-push job.
2. Uses the `aws-production` GitHub environment.
3. Has the explicit `push` + `aws` branch guard described above.
4. Authenticates to AWS through the existing GitHub OIDC role.
5. Sends the exact image URI to the EC2 deployment script through SSM:

   ```text
   745912973548.dkr.ecr.us-east-1.amazonaws.com/api-diafestivo:${GITHUB_SHA}-arm64
   ```

6. Waits for the SSM command to finish.
7. Prints non-sensitive command output.
8. Fails if the deployment or health check fails.

Add a deployment concurrency group such as `aws-production-deploy`, with in-progress deployments not running concurrently.

## 8. Server-side deployment behavior

The deployment script should:

1. Authenticate Docker to ECR using the EC2 instance role.
2. Pull the exact SHA-tagged image before stopping the current container.
3. Record the currently running image for rollback.
4. Replace the container using:
   - `--restart unless-stopped`
   - `--env-file /etc/api-diafestivo.env`
   - `-p 127.0.0.1:3002:3002`
5. Poll `http://127.0.0.1:3002/healthz` with a short timeout.
6. Finish successfully and clean up sufficiently old unused images if healthy.
7. Remove the failed container, restore the previous image, and fail the SSM command if unhealthy.

A simple replacement can produce a few seconds of Nginx `502` responses. Do not add blue/green deployment unless zero-downtime deployment becomes a requirement.

## 9. Verification

Before considering the setup complete, verify:

- A push to `main` cannot execute the deployment job.
- A manual workflow run cannot execute the deployment job.
- A push to `aws` deploys the image tagged with that exact commit SHA.
- `https://api.diafestivo.co/healthz` returns `200` and `ok`.
- The container restarts after an EC2 reboot.
- Port `3002` and SSH port `22` are unreachable publicly.
- An intentionally unhealthy image causes rollback and a failed GitHub Actions job.
- Certificate renewal is enabled.

## Expected repository changes during implementation

Keep the implementation small:

- Modify `.github/workflows/push-to-ecr.yaml`
- Add one `ops/bootstrap.sh` containing the one-time host setup, Nginx configuration, and deployment script installation
- Optionally add short operational instructions to `README.md`

Do not add Docker Compose, an SSH deployment action, Terraform, a load balancer, or another application health endpoint for this single-instance setup.

## Known limitations

- A single EC2 instance is not highly available.
- The host requires OS patching, disk monitoring, and backups where needed.
- Costs include EC2, EBS, public IPv4, DNS, and data transfer.
- Add an Application Load Balancer, Auto Scaling Group, or ECS only when availability or scaling requirements justify them.

## References

- [AWS Systems Manager Run Command](https://docs.aws.amazon.com/systems-manager/latest/userguide/run-command.html)
- [Pulling private Amazon ECR images](https://docs.aws.amazon.com/AmazonECR/latest/userguide/docker-pull-ecr-image.html)
- [GitHub Actions deployment concurrency](https://docs.github.com/actions/writing-workflows/choosing-what-your-workflow-does/control-the-concurrency-of-workflows-and-jobs)
