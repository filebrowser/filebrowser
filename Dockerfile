FROM node:alpine
WORKDIR /src/app
COPY . /src/app
RUN yarn install && rm -rf assets/dist && npm run build

FROM golang:alpine
WORKDIR /go/src/github.com/hacdias/filemanager
COPY . /go/src/github.com/hacdias/filemanager
COPY --from=0 /src/app/assets/dist /go/src/github.com/hacdias/filemanager/assets/dist
RUN apk add --no-cache git && go get -u github.com/golang/dep/cmd/dep && go get github.com/GeertJohan/go.rice/rice

RUN dep ensure -update && rice embed-go && cd ./caddy/hugo && rice embed-go
RUN cd /go/src/github.com/hacdias/filemanager/cmd/filemanager && go build

FROM alpine:latest
COPY --from=1 /go/src/github.com/hacdias/filemanager/cmd/filemanager/filemanager /usr/local/bin/filemanager
ENTRYPOINT ["filemanager"]
CMD ["-h"]