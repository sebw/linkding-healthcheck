FROM golang:1.22.9 as build
WORKDIR /app
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=build /app/main /usr/local/bin/main
CMD ["/usr/local/bin/main"]
