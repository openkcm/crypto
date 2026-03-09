GOEXPERIMENT=runtimesecret GOOS=linux GOARCH=amd64 go build .
docker run --privileged --rm --cap-add=IPC_LOCK --ulimit memlock=-1:-1 -it --name alpine -v "$(pwd)":/app -w /app golang:tip-trixie ../../../krypton
