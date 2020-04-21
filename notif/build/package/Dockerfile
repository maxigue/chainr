# Build stage
FROM golang:1.14.1-alpine as builder

WORKDIR /go/src
COPY . .

RUN go install -v

# Final container
FROM alpine:latest

WORKDIR /app

COPY --from=builder /go/bin/notif /app
COPY --from=builder /go/src/config.yaml /app/etc/config.yaml

ENV CONFIG_FILE /app/etc/config.yaml

EXPOSE 80

ENTRYPOINT ["./notif"]
