FROM golang:1.27.0-alpine AS build
ARG TARGETARCH
ARG TARGETOS
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -trimpath -ldflags="-s -w" -o api .

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=build /app/api .
USER nonroot:nonroot
EXPOSE 3002
CMD ["/app/api"]
