# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from the latest golang base image
FROM golang:latest

# Add Maintainer Info
LABEL maintainer="Stuart Davies"

# Set the Current Working Directory inside the container
WORKDIR /server

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY ./site ./site
COPY ./exec ./exec
COPY ./templates ./templates
COPY webServerExample webServerExample
COPY webServerExample.json webServerExample.json
COPY LICENSE LICENSE
COPY README.md README.md


# Build the Go app
##RUN go build -o webServerBase.go .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./webServerExample"]

