FROM golang:alpine

COPY . /go/src/github.com/hacdias/filemanager

WORKDIR /go/src/github.com/hacdias/filemanager
RUN apk add --no-cache git
RUN go get ./...

WORKDIR /go/src/github.com/hacdias/filemanager/cmd/filemanager
RUN CGO_ENABLED=0 go build -a
RUN mv filemanager /go/bin/filemanager

FROM scratch
COPY --from=0 /go/bin/filemanager /filemanager

VOLUME /srv
EXPOSE 80

COPY Docker.json /config.json

ENTRYPOINT ["/filemanager"]
CMD ["--config", "/config.json"]
