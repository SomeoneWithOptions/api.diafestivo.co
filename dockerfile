# Build stage
FROM golang:latest AS builder
WORKDIR /app/
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -o api

# Runtime stage
FROM ubuntu:latest
WORKDIR /app
COPY --from=builder /app/api .
COPY . .
CMD ["./api"]