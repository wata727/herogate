mock:
	mockgen -source vendor/github.com/aws/aws-sdk-go/service/codebuild/codebuildiface/interface.go -destination mock/codebuild.go -package mock
	mockgen -source vendor/github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface/interface.go -destination mock/cloudwatchlogs.go -package mock
	mockgen -source vendor/github.com/aws/aws-sdk-go/service/ecs/ecsiface/interface.go -destination mock/ecs.go -package mock

.PHONY: mock
