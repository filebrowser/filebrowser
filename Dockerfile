FROM golang:alpine

COPY . /go/src/github.com/filebrowser/filebrowser

WORKDIR /go/src/github.com/filebrowser/filebrowser
RUN apk add --no-cache git curl && \
  curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 && \
  chmod +x /usr/local/bin/dep
RUN dep ensure -vendor-only

WORKDIR /go/src/github.com/filebrowser/filebrowser/cmd/filebrowser
RUN CGO_ENABLED=0 go build -a
RUN mv filebrowser /go/bin/filebrowser

FROM scratch
COPY --from=0 /go/bin/filebrowser /filebrowser

VOLUME /tmp
VOLUME /srv
EXPOSE 80

COPY Docker.json /config.json

ENTRYPOINT ["/filebrowser", "--config", "/config.json"]
