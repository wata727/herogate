default: build

prepare:
	go get -u github.com/golang/dep/cmd/dep
	dep ensure

test: prepare
	go test $$(go list ./... | grep -v vendor | grep -v mock)

build: test
	go build -v

install: test
	go install

lint:
	go get -u github.com/client9/misspell/cmd/misspell
	golint -set_exit_status $$(go list ./... | grep -v vendor | grep -v mock)
	go vet $$(go list ./... | grep -v vendor | grep -v mock)
	misspell -error $$(find . -type f | grep -v vendor | grep -v mock)

mock:
	go get -u github.com/golang/mock/mockgen
	go generate ./...

assets:
	go get -u github.com/a-urth/go-bindata/...
	go generate ./...

.PHONY: default prepare test build install lint mock assets
