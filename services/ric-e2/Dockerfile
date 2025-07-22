# Build Stage
FROM golang:1.21 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/cmd/ric-e2
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ric-e2 .

# Final Stage
FROM gcr.io/distroless/static-debian11

WORKDIR /

COPY --from=builder /app/cmd/ric-e2/ric-e2 .
COPY --from=builder /app/config/ric-e2-config.yaml ./config.yaml

USER 65532:65532

EXPOSE 36421

ENTRYPOINT ["/ric-e2", "--config", "./config.yaml"]