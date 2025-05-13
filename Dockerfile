FROM golang:1.24-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o loadbalancer ./cmd

FROM golang:1.24-alpine
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/loadbalancer /app/loadbalancer
COPY . /app
COPY --from=builder /app/demo/start_mock_backends.sh /app/start_mock_backends.sh
COPY config.yaml .
RUN chmod +x /app/start_mock_backends.sh
EXPOSE 8080
CMD ["/bin/sh", "-c", "/app/start_mock_backends.sh && /app/loadbalancer"]