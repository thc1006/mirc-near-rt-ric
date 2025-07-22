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
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ric-a1 ./cmd/ric-a1

# Stage 2: Create the final image
FROM gcr.io/distroless/static-debian11

WORKDIR /

# Copy the binary from the builder stage
COPY --from=builder /app/ric-a1 .

# Add a non-root user
USER 65532:65532

# Expose the A1 port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/ric-a1"]