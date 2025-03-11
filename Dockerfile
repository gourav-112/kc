# Step 1: Build the Go app
FROM golang:1.24-alpine as builder

# Install gcc and musl-dev for building SQLite3 dependencies
RUN apk add --no-cache gcc musl-dev

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod tidy

# Copy the rest of the application code
COPY . .

# Build the Go app
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o main .

# Step 2: Create the final image
FROM alpine:latest

# Install ca-certificates for secure connections
RUN apk --no-cache add ca-certificates

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the compiled binary from the builder image
COPY --from=builder /app/main .

# Expose port 8080 for the application
EXPOSE 8080

# Command to run the application
CMD ["./main"]
