FROM golang:alpine AS base

RUN apk add -U --no-cache git && \
  go get github.com/alecthomas/gometalinter && \
  gometalinter --install

FROM golang:alpine

WORKDIR /go/src/github.com/filebrowser/filebrowser

COPY --from=base /go/bin /go/bin

RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories && \
  sed -i -e "s/v[0-9]\.[0-9]/edge/g" /etc/apk/repositories  && \
  apk --no-cache -U upgrade && apk --no-cache add ca-certificates yarn git curl dos2unix && \
  go get github.com/GeertJohan/go.rice/rice && \
  curl -fsSL -o /go/bin/dep $( \
    curl -s https://api.github.com/repos/golang/dep/releases/latest \
    | grep "browser_download_url.*linux-amd64\"" \
    | cut -d : -f 2,3 \
    | tr -d \" \
  ) && \
  chmod +x /go/bin/dep && \
  curl -sL https://git.io/goreleaser -o /go/bin/goreleaser && \
  chmod +x /go/bin/goreleaser && \
  curl -fsSL https://download.docker.com/linux/static/edge/x86_64/docker-18.05.0-ce.tgz | tar xvz --strip-components=1 docker/docker -C /go/bin && \
  chmod +x /go/bin/docker && \
  curl -fsSL $( \
    curl -s https://api.github.com/repos/docker/docker-credential-helpers/releases/latest \
    | grep "browser_download_url.*pass-.*-amd64"  \
    | cut -d : -f 2,3 \
    | tr -d \" \
  ) | tar xv -C /go/bin && \
  chmod + /go/bin/docker-credential-pass
