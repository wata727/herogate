FROM golang:1.9.4-alpine as build-layer

RUN apk add --no-cache git && go get -u github.com/golang/dep/cmd/dep

COPY . src/github.com/wata727/herogate/
WORKDIR src/github.com/wata727/herogate
RUN dep ensure && go install

FROM alpine:3.7

COPY --from=build-layer /go/bin/herogate /usr/local/bin

WORKDIR /workdir
ENTRYPOINT ["herogate"]
