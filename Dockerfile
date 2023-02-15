# Use a base image with Go and the Go module dependency manager installed
FROM golang:1.17-alpine as builder

# Set the working directory in the container
WORKDIR /app

# Copy the Go module files to the container
COPY go.mod .
COPY go.sum .

# Download the Go module dependencies
RUN go mod download

# Copy the source code to the container
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Use a minimal base image to reduce the image size
FROM alpine:latest

# Copy the built binary from the builder stage
COPY --from=builder /app/main /app/main
COPY --from=builder /app/index.html /app/index.html

# Set the working directory in the container
WORKDIR /app

# Expose port 8080 to the host
EXPOSE 8080

# Define the command to run when the container starts
# CMD ["/app/main"]

# Define the command to run when the container starts and show dir content
CMD ["/bin/sh", "-c", "ls -la . && /app/main"]