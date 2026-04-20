# Using a lightweight version of Go (Alpine Linux) as base image.
FROM golang:1.22-alpine

# The workfolder is placed inside the container to /app
WORKDIR /app

# Copying the main.go from the computer to inside the container
COPY main.go .

# Deactivating Go Modules since we're running a simple file without external libraries
ENV GO111MODULE=off

# Compiling code to a runnable binary with the name ids
RUN go build -o ids main.go

# Informing the container that it should listen to port 8080
EXPOSE 8080

# Running application when a container is started
CMD ["./ids"]
