FROM --platform=$BUILDPLATFORM golang:1.22.9 as build
WORKDIR /app
COPY . .
RUN go mod tidy
<<<<<<< Updated upstream
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o main .

||||||| Stash base
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .
=======
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o main .


>>>>>>> Stashed changes
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=build /app/main /usr/local/bin/main
CMD ["/usr/local/bin/main"]
