# ===========================
# Stage 1: Build the Go Binary
# ===========================
# Using the official Golang image for building the binary.
FROM golang:1.24.3 AS builder

# Create a non-root user (appuser) for building the application (better security).
RUN useradd --create-home appuser

# Set the working directory to the user's home directory.
WORKDIR /home/appuser/

# Copy go.mod and go.sum files for dependency management (optimized caching).
COPY go.mod go.sum ./

# Download the Go module dependencies (cached if no changes in go.mod or go.sum).
RUN go mod download

# Copy the application source code (separate directories for modular design).
COPY order/ ./order/
COPY common/ ./common/

# Change the working directory to the order service directory.
WORKDIR /home/appuser/order

# Build the Go binary for Linux (statically linked binary for better portability).
RUN CGO_ENABLED=0 GOOS=linux go build -o order-service .

# ===========================
# Stage 2: Minimal Runtime Image
# ===========================
# Using Distroless image (gcr.io/distroless/base-debian11) for a secure, minimal runtime.
FROM gcr.io/distroless/base-debian11

# Copy the compiled binary from the builder stage to the runtime image.
COPY --from=builder /home/appuser/order/order-service /usr/local/bin/order-service

# Set the application to run as a non-root user (increased security).
USER nonroot:nonroot

# Expose the HTTP port (8090) for the order service.
EXPOSE 8090

# Define the entrypoint command (starts the application).
ENTRYPOINT ["/usr/local/bin/order-service"]
