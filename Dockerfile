FROM golang:1.25.6-alpine3.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/app

FROM alpine:3.23

WORKDIR /app

COPY --from=builder /app/server .
COPY config/config.yaml ./config/

EXPOSE 8090

CMD ["./server"]
