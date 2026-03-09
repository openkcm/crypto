GOEXPERIMENT=runtimesecret GOOS=linux GOARCH=amd64 go build .
docker run --ulimit core=0 --rm -it --name alpine -v "$(pwd)":/app -w /app alpine ./cmd
