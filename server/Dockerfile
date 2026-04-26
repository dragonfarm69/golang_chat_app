# ==========================================
# STAGE 1: The Builder (Heavy)
# ==========================================
# Use the official Golang image to get all the compiler tools
FROM golang:1.26-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy your dependency files first (this caches them to speed up future builds)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of your actual code
COPY . .

# Compile the Go app into a single executable binary
RUN CGO_ENABLED=0 GOOS=linux go build -o chat-server .

# ==========================================
# STAGE 2: The Final Container (Ultra Lightweight)
# ==========================================
# Start over with a completely blank, tiny Linux container
FROM alpine:latest

WORKDIR /root/

# Copy ONLY the compiled binary from Stage 1 into this new container
COPY --from=builder /app/chat-server .

# Tell Docker this container will use port 8080
EXPOSE 8080

# Command to run when the container starts
CMD ["./chat-server"]