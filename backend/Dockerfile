FROM golang:1.23-alpine AS development

# Install development tools
RUN apk add --no-cache git curl

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Install a compatible version of air for hot reloading
RUN go install github.com/cosmtrek/air@v1.49.0

# Copy source code
COPY . .

# Expose port
EXPOSE 8080

# Use air for hot reloading in development
CMD ["air", "-c", ".air.toml"]