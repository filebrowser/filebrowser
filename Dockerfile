FROM golang:alpine
COPY . /go/src/github.com/hacdias/filemanager
WORKDIR /go/src/github.com/hacdias/filemanager
RUN apk add --no-cache git
RUN go get ./...
WORKDIR /go/src/github.com/hacdias/filemanager/cmd/filemanager
RUN go build
ENTRYPOINT ["/go/src/github.com/hacdias/filemanager/cmd/filemanager/filemanager"]
CMD ["-h"]
