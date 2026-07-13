# Simplification and Reorganization Plan

## Objective

Simplify the application, remove unused abstractions, and flatten the repository while preserving all user-visible behavior.

Implementation will happen only on the `dev` branch. Merging into `main` or `aws` is outside this plan.

## Compatibility contract

The refactor must not add, remove, or change any public endpoint or user-facing function.

Preserve exactly:

- HTTP methods and paths
- Status codes
- `Access-Control-Allow-Origin` headers
- Content types
- JSON field names, casing, values, and date formats
- Existing JSON newline behavior
- Error response text
- Rendered HTML
- Holiday calculations
- Environment variables
- External Giphy and IPInfo request behavior

### Routes

| Route | Current contract |
| --- | --- |
| `GET /all` | Current-year Colombian holidays as JSON |
| `GET /next` | Next holiday as JSON |
| `GET /template` | Current/next holiday HTML |
| `GET /is/{date}` | Holiday check as JSON; invalid dates return `400` and `error parsing date` |
| `GET /left` | Remaining-holidays HTML |
| `GET /make?year=YYYY` | Holidays for a year; invalid years return `400` and `error parsing year` |
| `GET /healthz` | `200`, plain text, body `ok` |
| Any unmatched route | Existing `400` JSON response with the same route list |

Unsupported-method behavior produced by the current `http.ServeMux` must also remain unchanged.

### Environment variables

Keep the names and behavior of:

- `PORT`
- `IP_INFO_TOKEN`
- `GIPHY_KEY`
- `MY_CIDR`

Internal Go packages, unreferenced functions, private types, and file locations are not part of this compatibility contract.

## Target repository structure

The application is one binary and does not need three internal packages. Move production code into one flat `main` package:

```text
.
├── .github/
│   └── workflows/
│       └── push-to-ecr.yaml
├── views/
│   ├── index.html
│   └── left.html
├── Dockerfile
├── README.md
├── SIMPLIFICATION_PLAN.md
├── go.mod
├── main.go
├── handlers.go
├── handlers_test.go
├── holidays.go
├── holidays_test.go
├── giphy.go
├── logging.go
└── logging_test.go
```

### File responsibilities

- `main.go`: configuration, logger setup, server construction, startup, and graceful shutdown.
- `handlers.go`: router, HTTP handlers, response helpers, embedded templates, and private view models.
- `holidays.go`: holiday data types, date calculations, sorting, and clock test seam.
- `giphy.go`: the single Giphy request and its minimal response model.
- `logging.go`: request logging middleware, response status recording, request IP parsing, CIDR checks, and IPInfo lookup.
- `views/`: unchanged template source files embedded into the binary.

Delete these directories after their code is moved:

- `giphy/`
- `holiday/`
- `templateinfo/`

## Implementation phases

### Phase 1: Prepare and protect `dev`

The local `dev` branch has been fast-forwarded to the current `main` branch. Do not merge the AWS-only deployment-plan commit into `dev` as part of this work.

Before implementation:

```sh
git switch dev
git status --short
go test ./...
```

Requirements:

- Worktree must be clean before refactoring.
- Keep each phase independently testable.
- Do not modify `main` or `aws`.

### Phase 2: Strengthen characterization tests

Use the existing `TestRouteContracts` test as the compatibility gate. Add only missing assertions needed to detect user-visible changes.

Add checks for:

1. Exact status, CORS header, and content type for every route.
2. Exact invalid-route JSON bytes, including its lack of a trailing newline.
3. Exact health response body.
4. Existing newline behavior for responses produced by `writeJSON`.
5. Existing invalid date and invalid year text.
6. Existing HTML content for `/template` and `/left` under a fixed clock.
7. Existing behavior for an unsupported method and an unmatched path.
8. All known fixed, Monday-shifted, Easter-derived, and post-2026 holiday dates.

Do not introduce snapshots, fixtures, golden-file frameworks, or new dependencies. Continue using `httptest` and the standard library.

Verification:

```sh
gofmt -w .
go test ./...
```

### Phase 3: Shrink the Giphy client

Current problem: `giphy/types.go` mirrors more than 200 lines of the Giphy response while the application reads only one value.

Implementation:

1. Move `giphy/gif.go` to root `giphy.go` and change it to `package main`.
2. Replace `Gif` with a private minimal response struct containing only:
   - `data`
   - `images`
   - `original`
   - `url`
3. Keep the request URL, API key, `celebrate` tag, `g` rating, context, client timeout, response-size limit, status validation, and missing-URL error unchanged.
4. Remove the unused context-free `FetchGifURL` wrapper.
5. Update the template handler to call the remaining context-aware function directly.
6. Delete `giphy/types.go` and then the empty `giphy/` directory.

No live Giphy request should be added to the tests.

### Phase 4: Remove unused IPInfo surface

Current problem: IPInfo Lite and context-free methods are unreferenced, while a generic helper supports APIs the application does not use.

Implementation:

1. Move the retained IPInfo logic into root `logging.go`.
2. Keep only the response fields written to request logs:
   - `ip`
   - `city`
   - `region`
   - `country`
3. Remove:
   - `IPInfoLite`
   - `FetchIPInfo`
   - `FetchIPInfoLite`
   - `FetchIPInfoLiteContext`
   - generic host/base-path/target parameters used only by the removed Lite client
4. Replace the `IP` wrapper type with small private functions where that reduces code:
   - context-aware IPInfo fetch
   - CIDR containment check
5. Preserve IP validation, token validation, URL encoding, response-size limit, HTTP timeout, non-2xx handling, and context cancellation.
6. Move the current CIDR tests into `logging_test.go` without changing their cases.

### Phase 5: Simplify holiday values and remove caching

Current problem: computing 18–19 dates is wrapped in a global cache, cloning, pointer-returning slice APIs, and a cache-only test.

Implementation:

1. Move `holiday/holiday.go` and retained types from `holiday/types.go` into root `holidays.go` with `package main`.
2. Remove `sync.Map` and all clone-on-load/store logic.
3. Generate and sort the small holiday list on demand.
4. Change internal APIs to return values:
   - `MakeHolidaysByYear(year int) Holidays`
   - `GetRemaining() Holidays`
   - `FindUpcomingHoliday() NextHoliday`
5. Use value receivers for read-only `Holidays` methods where practical.
6. Remove unnecessary constructors and construct `Holiday` and `NextHoliday` values directly.
7. Remove unused `NewHoliday` and `MakeDatesInCOT` functions.
8. Keep the clock test seam, Colombia fixed timezone, Easter algorithm, Monday shifting, date comparison, names, sorting, and 2026 Chiquinquirá rule unchanged.
9. Delete the cache-cloning test because the cache no longer exists.
10. Move retained holiday tests to root `holidays_test.go`.
11. Delete the empty `holiday/` directory.

Run the holiday tests before updating handlers, then run the complete suite after callers use value returns.

### Phase 6: Consolidate handlers, responses, and templates

Move these files into root `handlers.go`:

- `routes.go`
- `handlers_json.go`
- `handlers_html.go`
- `responses.go`
- `templates.go`

Implementation details:

1. Keep `newServeMux` and every current route pattern unchanged.
2. Remove the unused exported route wrappers:
   - `HandleAllRoute`
   - `HandleNextRoute`
   - `HandleTemplateRoute`
   - `HandleIsRoute`
   - `LeftHandler`
   - `MakeHandler`
   - `HandleInvalidRoute`
3. Replace `templateinfo.TemplateInfo` and its nine-argument constructor with a private view struct and a keyed struct literal.
4. Keep the current private `/left` view struct private and local unless sharing it reduces code.
5. Delete the `templateinfo/` directory.
6. Keep the current JSON, text, HTML, and CORS behavior.

#### Invalid-route response

Remove:

- `invalidRouteResponse` startup-only type
- `init()` marshalling
- mutable `invalidRouteResponseBody`
- `writeJSONBytes`

Use a static JSON string and write it directly in `handleInvalidRoute`. The bytes must remain exactly:

```json
{"status":400,"message":"Please Use Valid Routes:","valid_routes":["/all","/next","/is/YYYY-MM-DD","/make?year=YYYY"]}
```

Do not route this response through `json.Encoder`, because that would add a trailing newline and change the current response bytes.

#### Embedded templates

Use standard-library `go:embed` for `views/index.html` and `views/left.html`.

Requirements:

- Do not edit template contents during this refactor.
- Parse templates once at startup.
- Keep buffered execution so template errors do not send partial successful responses.
- Keep rendered HTML byte-for-byte equivalent.
- Keep the source files under `views/` for editing.

### Phase 7: Simplify configuration and startup

Move `config.go` into `main.go`.

Implementation:

1. Preserve port defaulting and validation.
2. Preserve all server and shutdown timeout values.
3. Preserve graceful shutdown behavior and exit handling.
4. Remove the unused `giphyKey` configuration field; the Giphy client already reads `GIPHY_KEY` at request time.
5. Keep only configuration consumed by the server or request logger.
6. Do not make fixed timeout constants configurable unless a current environment variable already controls them.

### Phase 8: Simplify logging

Move `request_logging.go` and retained `ipinfo.go` logic into root `logging.go`.

Implementation:

1. Delete `level_split_handler.go`.
2. Configure one `slog.NewJSONHandler` instead of implementing the complete `slog.Handler` interface solely to split stdout and stderr.
3. Keep structured log levels so the container log collector can filter errors.
4. Remove `requestLogData.loggedTime` and the custom formatted `time` attribute; the JSON handler already writes a timestamp.
5. Remove the logging package's dependency on holiday time functions.
6. Preserve:
   - asynchronous logging after the response
   - logging timeout
   - detached request context with deadline
   - response status capture
   - first `X-Forwarded-For` value behavior
   - `RemoteAddr` fallback
   - `MY_CIDR` exclusion
   - one-time invalid-CIDR error
   - local fallback logging
   - IPInfo enrichment
   - cancellation-error suppression
   - method, URL, path, protocol, status, duration, IP, city, region, and country fields
7. Merge `ipinfo_test.go` and `request_logging_test.go` into `logging_test.go`.

Logging format changes are limited to removing the redundant custom COT timestamp and using the standard structured timestamp. HTTP responses remain unaffected.

### Phase 9: Simplify Docker and CI

#### GitHub Actions

In `.github/workflows/push-to-ecr.yaml`:

1. Keep `Setup Go` in the test job.
2. Remove `Setup Go` from both image-build jobs because the Docker builder compiles the application.
3. Preserve test, formatting, vet, `go fix`, and staticcheck checks.
4. Preserve runners, AWS authentication, repository name, architectures, image tags, and pushes.
5. Do not add a matrix, reusable workflow, or custom action.

#### Docker

After templates are embedded and the module remains standard-library-only:

1. Remove the runtime `COPY /app/views/` instruction.
2. Remove the empty `go.sum` file.
3. Replace the no-op `go mod download` layer with the existing source copy and build.
4. Preserve:
   - multi-stage build
   - target architecture support
   - static binary flags
   - distroless runtime
   - non-root user
   - port `3002`
   - `/app/api` command

Clean obsolete `.dockerignore` entries only when their referenced files no longer exist.

## Audit finding traceability

| Audit finding | Planned implementation |
| --- | --- |
| Oversized Giphy schema | Phase 3 minimal private response struct |
| Unused compatibility/client APIs | Phases 3, 4, 5, and 6 remove them |
| Custom split logger | Phase 8 uses one JSON handler |
| Pre-marshalled invalid response | Phase 6 uses an exact static response |
| Holiday cache and clones | Phase 5 computes on demand |
| Single-consumer `templateinfo` package | Phase 6 private keyed view struct |
| Pointer-returning holiday APIs | Phase 5 value returns |
| Unused Go setup in Docker jobs | Phase 9 removes both steps |
| Duplicate request timestamp | Phase 8 relies on `slog` timestamp |

## Commit sequence

Keep changes reviewable and the test suite green after every commit:

1. `test: lock public http behavior`
2. `refactor: shrink giphy and ipinfo clients`
3. `refactor: simplify holiday calculations`
4. `refactor: consolidate handlers and embed views`
5. `refactor: flatten application packages`
6. `refactor: simplify startup and request logging`
7. `chore: simplify docker and ci`

File moves and their package/import updates should be committed together so no commit leaves the repository unbuildable.

## Verification after every phase

```sh
gofmt -w .
go test ./...
go vet ./...
```

## Final verification

```sh
gofmt -w .
go test ./...
go vet ./...
go fix -diff ./...
docker build -t api-diafestivo:dev .
docker run --rm -p 3002:3002 api-diafestivo:dev
```

Smoke-test:

```sh
curl -i http://127.0.0.1:3002/all
curl -i http://127.0.0.1:3002/next
curl -i http://127.0.0.1:3002/is/2025-01-01
curl -i http://127.0.0.1:3002/is/bad
curl -i 'http://127.0.0.1:3002/make?year=2027'
curl -i http://127.0.0.1:3002/template
curl -i http://127.0.0.1:3002/left
curl -i http://127.0.0.1:3002/healthz
curl -i http://127.0.0.1:3002/not-a-route
```

Compare all responses against the characterization tests. No implementation phase is complete if an endpoint, header, response body, holiday date, template result, or environment-variable behavior changes.

## Completion criteria

- [ ] Work exists only on `dev`.
- [ ] All current endpoints and user-visible behavior are preserved.
- [ ] No new dependency is added.
- [ ] `giphy/`, `holiday/`, and `templateinfo/` are removed.
- [ ] Production code is organized into the five focused root Go files.
- [ ] Giphy decoding retains only the required URL field.
- [ ] Unreferenced compatibility APIs are removed.
- [ ] Holiday caching, cloning, and pointer-to-slice APIs are removed.
- [ ] Templates are embedded without output changes.
- [ ] Standard `slog` replaces the custom split handler.
- [ ] Docker image jobs no longer install Go on the runner.
- [ ] Docker runtime contains only the compiled application artifact.
- [ ] Formatting, tests, vet, `go fix`, Docker build, and endpoint smoke tests pass.
- [ ] Nothing is merged into `main` or `aws` without separate approval.
