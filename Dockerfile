FROM golang:1.24-alpine AS builder
ENV GOTOOLCHAIN=auto

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/service ./cmd/service/main.go

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -g 10001 appgroup && \
    adduser -u 10001 -G appgroup -S -D appuser

USER appuser

COPY --from=builder /app/bin/service /app/service

EXPOSE 8080

CMD ["./service"]