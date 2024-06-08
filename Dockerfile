FROM golang:1.22-alpine AS build
WORKDIR /app/
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=build /app/api . 
COPY /index.html .
COPY /en.html .

CMD ["./api"]
