.PHONY: clean
clean:
	rm -f cover.out cover.html krypton
	rm -rf cover/

.PHONY: lint
lint:
	golangci-lint run --fix ./...


.PHONY: test
test: clean
	@mkdir -p cover/integration cover/unit
	@go clean -testcache

	go test -count=1 -race -cover ./... -args -test.gocoverdir="${PWD}/cover/unit"
	GOCOVERDIR="${PWD}/cover/integration" go test -count=1 -race --tags=integration ./integration

	@go tool covdata textfmt -i=./cover/unit,./cover/integration -o cover.out

	@echo "On a Mac, you can use the following command to open the coverage report in the browser\ngo tool cover -html=cover.out -o cover.html && open cover.html"

