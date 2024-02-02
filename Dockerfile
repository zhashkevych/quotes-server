# Build stage
FROM golang:1.21.5-alpine3.18 AS build

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o client ./cmd/client/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go

# Final stage
FROM alpine:latest
WORKDIR /root/

COPY --from=build /app/client .
COPY --from=build /app/server .