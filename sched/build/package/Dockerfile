# Build stage
FROM golang:1.14.1-alpine as builder

WORKDIR /go/src
COPY . .

RUN go install -v

# Final container
FROM alpine:latest

WORKDIR /app

COPY --from=builder /go/bin/sched /app

EXPOSE 8080

ENTRYPOINT ["./sched"]
