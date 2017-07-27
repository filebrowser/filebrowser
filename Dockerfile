FROM golang:alpine
WORKDIR /go/src/github.com/hacdias/filemanager
COPY . /go/src/github.com/hacdias/filemanager

RUN apk add --no-cache git && go get ./... && cd /go/src/github.com/hacdias/filemanager/cmd/filemanager && go build

FROM alpine:latest
COPY --from=0 /go/src/github.com/hacdias/filemanager/cmd/filemanager/filemanager /usr/local/bin/filemanager
ENTRYPOINT ["filemanager"]
CMD ["-h"]