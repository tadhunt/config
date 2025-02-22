export PROJECT := cleverlikelms-dev
export GOOGLE_APPLICATION_CREDENTIALS := ./secrets/credentials.json

all:
	go mod tidy
	go vet
	staticcheck
	go build

test: all
	@if [ -z "${GOOGLE_APPLICATION_CREDENTIALS}" ] ; then echo "must set GOOGLE_APPLICATION_CREDENTIALS"; exit 1; fi
	go test -failfast -count=1 -v ./...
