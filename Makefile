all:
	go mod tidy
	go vet
	staticcheck
	go build

test:
	PROJECT=test-project go test -failfast -count=1 -v ./...
