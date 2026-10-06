# Stage 1: Build binary
FROM docker.io/library/golang:1.27-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /build/bin/app ./cmd/app
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /build/bin/client ./cmd/client

# Stage 2: Runtime image
FROM docker.io/library/alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /build/bin/app /app/app
COPY --from=builder /build/bin/client /app/client
COPY --from=builder /build/migrations /app/migrations

ENV PATH="/app:${PATH}"

EXPOSE 50051 9090

CMD ["./app"]
