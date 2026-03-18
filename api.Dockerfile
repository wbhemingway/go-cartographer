# Build Stage

FROM golang:1.25.4-alpine3.21 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /api-server ./cmd/api

# Deploy Stage

FROM alpine:3.21

WORKDIR /

COPY --from=builder /api-server /api-server

EXPOSE 8080

ENTRYPOINT ["/api-server"]
