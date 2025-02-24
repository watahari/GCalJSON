# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o gcaljson

# Final stage using Alpine for CA certificates support
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/gcaljson /gcaljson
EXPOSE 8080
ENTRYPOINT ["/gcaljson"]
