FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o trip-service ./cmd


FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/trip-service .

EXPOSE 50053

CMD ["./trip-service"]