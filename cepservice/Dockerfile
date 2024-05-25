# Dockerfile
FROM golang:1.22.3-alpine

WORKDIR /app/cepservice

COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -C ./cmd/cepsystem -o main

CMD ["./cmd/cepsystem/main"]
