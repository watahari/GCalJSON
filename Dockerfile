# Build stage
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o gcaljson

# Final stage
FROM scratch
COPY --from=builder /app/gcaljson /gcaljson
EXPOSE 8080
ENTRYPOINT ["/gcaljson"]
