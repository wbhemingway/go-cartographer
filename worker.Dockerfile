# Build Stage

FROM golang:1.25.4-alpine3.21 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /worker ./cmd/worker

# Deploy Stage

FROM alpine:3.21

WORKDIR /

COPY --from=builder /worker /worker

COPY --from=builder /app/assets /assets

ENTRYPOINT ["/worker"]
