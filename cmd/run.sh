GOOS=linux GOARCH=amd64 go build .
docker run --rm -it --name alpine -v "$(pwd)":/app -w /app alpine ./cmd

