FROM golang:alpine

COPY . /go/src/github.com/hacdias/filemanager

WORKDIR /go/src/github.com/hacdias/filemanager
RUN apk add --no-cache git
RUN go get ./...

WORKDIR /go/src/github.com/hacdias/filemanager/cmd/filemanager
RUN go install

VOLUME /srv
EXPOSE 80

COPY Docker.json /etc/config.json

ENTRYPOINT ["/go/bin/filemanager"]
CMD ["--config", "/etc/config.json"]
