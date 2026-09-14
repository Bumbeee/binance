FROM golang:1.26-alpine AS builder

WORKDIR /src
COPY proto/ ./proto/
COPY shared/ ./shared/
COPY spot-instrument-service/ ./spot-instrument-service/

WORKDIR /src/spot-instrument-service
ENV GOWORK=off
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/spot-instrument-service ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /out/spot-instrument-service /usr/local/bin/spot-instrument-service
RUN chmod +x /usr/local/bin/spot-instrument-service
EXPOSE 50052
ENTRYPOINT ["/usr/local/bin/spot-instrument-service"]