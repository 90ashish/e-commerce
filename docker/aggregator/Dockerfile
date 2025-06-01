# --- Stage 1: Build the Go binary ---
FROM golang:1.24.3 AS builder

# 1) Set working directory inside the container
WORKDIR /src

# 2) Copy the *root* go.mod & go.sum so "module e-commerce" is visible
COPY go.mod go.sum ./

# 3) Copy the aggregator’s own go.mod & go.sum
COPY aggregator/go.mod aggregator/go.sum ./aggregator/

# 4) Copy the local 'common/' folder so the replace directive finds it
COPY common/ ./common/

# 5) Copy the aggregator source code
COPY aggregator/ ./aggregator/

# 6) Now that all go.mod files and source code are in place, download deps
RUN go mod download

# 7) Build the binary (disable CGO for portability)
WORKDIR /src/aggregator
RUN CGO_ENABLED=0 GOOS=linux go build -o aggregator-service .

# --- Stage 2: Minimal runtime image ---
FROM gcr.io/distroless/base-debian11

# Copy the compiled binary from the builder stage
COPY --from=builder /src/aggregator/aggregator-service /usr/local/bin/aggregator-service

# Run as non-root user
USER nonroot:nonroot

# Entrypoint to run the service
ENTRYPOINT ["/usr/local/bin/aggregator-service"]
