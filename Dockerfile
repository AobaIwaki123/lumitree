# Build stage (Native Go cross-compilation for ultra-fast multi-arch builds)
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code (filtered via .dockerignore)
COPY . .

# Build static binary for target architecture
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
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
