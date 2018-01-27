mock:
	mockgen -source vendor/github.com/aws/aws-sdk-go-v2/service/codebuild/codebuildiface/interface.go -destination mock/codebuildmock.go -package mock

.PHONY: mock
