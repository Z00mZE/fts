#   Builder-container
FROM golang:alpine as builder
RUN apk add tzdata git ca-certificates \
    && cp /usr/share/zoneinfo/Europe/Moscow /etc/localtime \
    && echo "Europe/Moscow" >  /etc/timezone \
    && date \
    && apk del tzdata
WORKDIR /build
COPY . .
RUN go mod download  \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o service ./cmd/main.go

#   Compact application container
FROM alpine:latest
COPY --from=builder /build/service /service
COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /zoneinfo.zip
ENV TZ=Europe/Moscow
ENTRYPOINT ["/service"]

