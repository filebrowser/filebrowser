# Build the frontend
FROM node:latest AS frontend-builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# 1. Remove old build files (if present)
# 2. Set the current directory to frontend
# 3. Install dependencies
# 4. Build the frontend
RUN rm -rf frontend/dist && \
    rm -rf frontend/node_modules && \
    rm -f http/rice-box.go && \
    cd /app/frontend && \
    npm install && \
    npm run build


# Build the binary
FROM golang:alpine as binary-builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Copy the source from the current directory to the Working Directory inside the container
COPY --from=frontend-builder /app .

# 1. Install requirements
# 2. Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
# 3. Install dependencies
# 4. Build the Go app
RUN apk add --no-cache gcc musl-dev && \
    go mod download && \
    go install github.com/GeertJohan/go.rice/rice && \
    cd /app/http && \
    rm -rf rice-box.go && \
    rice embed-go && \
    cd /app && \
    go build -a -o filebrowser -ldflags "-s -w"


# Add CA certificates
FROM alpine:latest AS certifier

# Install CA certificates
RUN apk --update add ca-certificates


# Serve app
FROM alpine
COPY --from=certifier /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=binary-builder /app/filebrowser /filebrowser

VOLUME /srv
EXPOSE 80

COPY .docker.json /.filebrowser.json

ENTRYPOINT [ "/filebrowser" ]
