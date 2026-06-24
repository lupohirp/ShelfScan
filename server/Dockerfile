FROM golang:1.26-alpine AS builder

WORKDIR /app

# SQLite requires CGo, install compiler and headers
RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -v -ldflags='-s -w' -trimpath -o main ./cmd/main.go

FROM alpine:latest

# Install runtime dependencies for SQLite if any
RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/clienti.csv /app/clienti.csv
COPY --from=builder /app/agenti.csv /app/agenti.csv
COPY --from=builder /app/db /app/db
COPY --from=builder /app/db/migrations /app/migrations

EXPOSE 8080

CMD ["./main"]
