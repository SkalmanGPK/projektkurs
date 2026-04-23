# Using a lightweight version of Go (Alpine Linux) as base image.
FROM golang:1.22-alpine

# The workfolder is placed inside the container to /app
WORKDIR /app

# Copying the main.go from your computer to the container
COPY main.go .

# Deactivating Go Modules since we're using only standard libraries
ENV GO111MODULE=off

# Compiling the code. 
# Ensuresthe binary is statically linked (best for Alpine)
RUN CGO_ENABLED=0 go build -o app2-proxy main.go

# Informing the container to listen on port 8080
EXPOSE 8080

# Running the application
CMD ["./app2-proxy"]
