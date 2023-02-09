# Use a base image with Go installed
FROM golang:latest

# Set the working directory in the container
WORKDIR /app

# Copy the source code to the container
COPY . .

RUN cat /etc/resolv.conf
# Build the Go application
RUN go build -o main .

# Expose port 8080 to the host
EXPOSE 8080

# Define the command to run when the container starts
CMD ["./main"]