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
	mockgen -source api/iface/client.go -destination mock/client.go -package mock
	mockgen -source vendor/github.com/aws/aws-sdk-go/service/codebuild/codebuildiface/interface.go -destination mock/codebuild.go -package mock
	mockgen -source vendor/github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface/interface.go -destination mock/cloudwatchlogs.go -package mock
	mockgen -source vendor/github.com/aws/aws-sdk-go/service/ecs/ecsiface/interface.go -destination mock/ecs.go -package mock

.PHONY: default prepare test build install lint mock
