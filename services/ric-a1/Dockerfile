# Build Stage
FROM golang:1.21 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/cmd/ric-a1
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ric-a1 .

# Final Stage
FROM gcr.io/distroless/static-debian11

WORKDIR /

COPY --from=builder /app/cmd/ric-a1/ric-a1 .
COPY --from=builder /app/config/ric-a1-config.yaml ./config.yaml

USER 65532:65532

EXPOSE 8080

ENTRYPOINT ["/ric-a1", "--config", "./config.yaml"]