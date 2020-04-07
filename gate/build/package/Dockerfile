# Build stage
FROM golang:1.14.1-alpine as builder

WORKDIR /go/src
COPY . .

RUN go install -v

# Final container
FROM alpine:latest

WORKDIR /root

COPY --from=builder /go/bin/gate .

EXPOSE 8080

ENTRYPOINT ["./gate"]
