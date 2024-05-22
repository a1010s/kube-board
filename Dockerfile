# Use a specific Go version as the base image
FROM golang:1.17-alpine AS build

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source code into the container
COPY . .

# Initialize Go module
RUN go mod init kube-board
RUN go mod tidy

# Build the Go app with necessary flags
RUN go build -o kube-board .

# Start a new stage from scratch
FROM alpine:latest

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=build /app/kube-board .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./kube-board"]
