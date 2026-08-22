# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /lumitree ./cmd/lumitree

# Runtime stage (Distroless for ultra-lightweight and secure container)
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

COPY --from=builder /lumitree /lumitree

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/lumitree"]
CMD ["serve", "--port", "8080", "--host", "0.0.0.0"]
