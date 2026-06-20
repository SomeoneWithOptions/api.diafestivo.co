FROM golang:1.26.4-alpine AS build
ARG TARGETARCH
ARG TARGETOS
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -trimpath -ldflags="-s -w" -o api .

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=build /app/api .
COPY --from=build /app/views/ ./views/
USER nonroot:nonroot
EXPOSE 3002
CMD ["/app/api"]
