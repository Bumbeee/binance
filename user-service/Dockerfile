FROM golang:1.26-alpine AS builder

WORKDIR /src
COPY proto/ ./proto/
COPY user-service/ ./user-service/

WORKDIR /src/user-service
ENV GOWORK=off
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/user-service ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /out/user-service /usr/local/bin/user-service
RUN chmod +x /usr/local/bin/user-service
EXPOSE 50051
ENTRYPOINT ["/usr/local/bin/user-service"]