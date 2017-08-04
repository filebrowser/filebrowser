FROM golang:alpine

COPY . /go/src/github.com/hacdias/filemanager

WORKDIR /go/src/github.com/hacdias/filemanager
RUN apk add --no-cache git
RUN go get ./...

WORKDIR /go/src/github.com/hacdias/filemanager/cmd/filemanager
RUN go build -ldflags "-X main.version=$(git tag -l --points-at HEAD)"
RUN mv filemanager /go/bin/filemanager

FROM alpine:latest
COPY --from=0 /go/bin/filemanager /usr/local/bin/filemanager

VOLUME /srv
EXPOSE 80

COPY Docker.json /etc/config.json

ENTRYPOINT ["/usr/local/bin/filemanager"]
CMD ["--config", "/etc/config.json"]
