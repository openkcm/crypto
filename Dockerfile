# syntax=docker/dockerfile:1

FROM golang:1.26 AS builder

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/reference/dockerfile/#copy
COPY . ./

# Build
 

RUN go build -C ./cmd/new2/ -o ss

# Optional:
# To bind to a TCP port, runtime parameters must be supplied to the docker command.
# But we can document in the Dockerfile what ports
# the application is going to listen on by default.
# https://docs.docker.com/reference/dockerfile/#expose

RUN cp ./cmd/new2/ss .
RUN ls -R


FROM alpine:latest  


WORKDIR /home/appuser

# Copy the binary from the builder stage
COPY --from=builder /app/ss .


# Run
CMD ["./ss"]
