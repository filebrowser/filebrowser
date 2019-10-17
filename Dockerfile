# Build the frontend
FROM node:latest AS frontend-builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Remove old build files (if present)
RUN rm -rf frontend/dist && \
    rm -f http/rice-box.go

# Set the Current Working Directory inside the container
WORKDIR /app/frontend

# Install dependencies
RUN npm install

# Build the frontend
RUN npm run build


# Build the binary
FROM golang:alpine as binary-builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Install requirements
RUN apk add --no-cache gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY --from=frontend-builder /app .

# Install dependencies
RUN go install github.com/GeertJohan/go.rice/rice
RUN cd /app/http && \
    rm -rf rice-box.go && \
    rice embed-go && \
    cd /app

# Build the Go app
RUN go build -a -o filebrowser -ldflags "-s -w"


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
