FROM golang:1.26-alpine AS builder

WORKDIR /app

# Gateway module
COPY vital_gateway/go.mod vital_gateway/go.sum ./vital_gateway/

# Service module metadata + protobuf packages
COPY auth_service/go.mod ./auth_service/
COPY auth_service/proto ./auth_service/proto

COPY payment_service/go.mod ./payment_service/
COPY payment_service/proto ./payment_service/proto

COPY trip_service/go.mod ./trip_service/
COPY trip_service/proto ./trip_service/proto

WORKDIR /app/vital_gateway

RUN go mod download

COPY vital_gateway/ .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o vital-gateway ./cmd/server


FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/vital_gateway/vital-gateway .

EXPOSE 8080

CMD ["./vital-gateway"]