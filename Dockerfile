FROM golang:1.22-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o frontend_app ./cmd/web
RUN CGO_ENABLED=0 go build -o auth_app ./cmd/api

RUN echo '#!/bin/sh' > start.sh && \
    echo './auth_app &' >> start.sh && \
    echo './frontend_app' >> start.sh && \
    chmod +x start.sh

EXPOSE 8080 8081

CMD ["./start.sh"]