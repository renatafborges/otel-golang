FROM golang:1.22.3-alpine

WORKDIR /app/tempservice

COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -C ./cmd/tempsystem -o main

CMD ["./cmd/tempsystem/main"]
