# build stage
FROM golang:1.23.6-alpine3.21 AS builder

RUN adduser -D -u 1001 gouser

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -tags whatsauto -o \
    /app/whatsauto_api \
    ./cmd/api

# Final image
FROM scratch

WORKDIR /app

COPY --from=builder /etc/passwd /etc/passwd

COPY --from=builder /app/whatsauto_api /app/whatsauto_api

COPY --from=builder /app/.env /app/.env

USER gouser

EXPOSE 8080

CMD ["./whatsauto_api"]