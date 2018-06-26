FROM golang:alpine

ARG FBURL=https://github.com/filebrowser/filebrowser
ARG DPURL=https://api.github.com/repos/golang/dep/releases/latest

RUN apk add --no-cache git curl

WORKDIR /go/src/github.com/filebrowser
RUN git clone ${FBURL}
# COPY . /go/src/github.com/filebrowser/filebrowser

WORKDIR /go/src/github.com/filebrowser/filebrowser
RUN curl -fsSL "$(curl -s "${DPURL}" \
  | grep -i 'browser_download_url.*linux-amd64"' \
  | cut -d '"' -f 4)" -o /usr/local/bin/dep \
 && chmod +x /usr/local/bin/dep
RUN dep ensure -vendor-only

WORKDIR /go/src/github.com/filebrowser/filebrowser/cmd/filebrowser
RUN CGO_ENABLED=0 go build -a
RUN mv filebrowser /go/bin/filebrowser

FROM scratch
COPY --from=0 /go/bin/filebrowser /filebrowser

COPY --from=0 /go/src/github.com/filebrowser/filebrowser/Docker.json /config.json
# COPY Docker.json /config.json

VOLUME /tmp
VOLUME /srv
EXPOSE 80

ENTRYPOINT ["/filebrowser", "--config", "/config.json"]
