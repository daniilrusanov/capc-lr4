FROM golang:1.18

WORKDIR /app

COPY . .

RUN go mod tidy

CMD ["go", "run", "cmd/web/main.go"]
