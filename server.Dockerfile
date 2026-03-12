# Build Stage

FROM golang:1.25.4-alpine3.21 AS Builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /server ./cmd/server

# Deploy Stage

FROM alpine:3.21

WORKDIR /

COPY --from=Builder /server /server

COPY --from=Builder /app/assets /assets

EXPOSE 8080

ENTRYPOINT ["/server"]
