# Step 1: Build the Go binary
FROM golang:1.23-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Go Modules manifests
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum are not changed
RUN go mod tidy

# Copy the source code into the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o newsbot .

# Step 2: Create the final image using a smaller base image
FROM alpine:latest  

# Install necessary certificates for HTTPS requests (if needed)
RUN apk --no-cache add ca-certificates

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the pre-built binary from the builder stage
COPY --from=builder /app/newsbot .

# Copy the configuration file into the container
COPY config/webconfig.yaml /root/config/webconfig.yaml

# Expose the port your application runs on
EXPOSE 8080

# Command to run the application
CMD ["./newsbot"]
