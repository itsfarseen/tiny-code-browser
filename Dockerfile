# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY main.go template.html ./

RUN go build -o main main.go

# Final stage
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .

ENV ROOT_DIR=/srv
ENV LISTEN_ADDR=0.0.0.0:80
ENV APP_TITLE="Tiny Code Browser"

RUN mkdir -p /srv
EXPOSE 80

CMD ["./main"]
