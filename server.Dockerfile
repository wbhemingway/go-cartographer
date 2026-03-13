# Build Stage

FROM golang:1.25.4-alpine3.21 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /server ./cmd/server

# Deploy Stage

FROM alpine:3.21

WORKDIR /

COPY --from=builder /server /server

COPY --from=builder /app/assets /assets

EXPOSE 8080

ENTRYPOINT ["/server"]
