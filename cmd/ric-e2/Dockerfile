# Stage 1: Build the Go binary
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ric-e2 ./cmd/ric-e2

# Stage 2: Create the final image
FROM gcr.io/distroless/static-debian11

WORKDIR /

# Copy the binary from the builder stage
COPY --from=builder /app/ric-e2 .

# Add a non-root user
USER 65532:65532

# Expose the E2 port
EXPOSE 36421

# Run the binary
ENTRYPOINT ["/ric-e2"]