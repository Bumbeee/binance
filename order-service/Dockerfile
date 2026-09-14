FROM golang:1.26-alpine AS builder

WORKDIR /src
COPY proto/ ./proto/
COPY shared/ ./shared/
COPY order-service/ ./order-service/

WORKDIR /src/order-service
ENV GOWORK=off
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/order-service ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /out/order-service /usr/local/bin/order-service
RUN chmod +x /usr/local/bin/order-service
EXPOSE 50053
ENTRYPOINT ["/usr/local/bin/order-service"]