FROM --platform=$BUILDPLATFORM golang:1.22.9 as build
WORKDIR /app
COPY . .
RUN go mod tidy
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=build /app/main /usr/local/bin/main
CMD ["/usr/local/bin/main"]
