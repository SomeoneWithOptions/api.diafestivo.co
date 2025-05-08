FROM golang:alpine AS build
ARG TARGETARCH
ARG TARGETOS
WORKDIR /app/
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=${TARGETOS-linux} GOARCH=${TARGETARCH-amd64} go build -o api .

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=build /app/api .
COPY /views/ /app/views/
CMD ["/app/api"]
