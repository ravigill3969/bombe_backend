# ---------- Build ----------
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o payment-service ./cmd/server


# ---------- Runtime ----------
FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/payment-service .

EXPOSE 50051

CMD ["./payment-service"]