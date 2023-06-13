FROM node:14.21-slim as nbuild
WORKDIR /app
COPY  ./src/frontend ./
RUN npm i
RUN npm run build

FROM golang:alpine as base
WORKDIR /app
COPY  ./src/backend ./
RUN go build -ldflags='-w -s' -o filebrowser .

FROM alpine:latest
RUN apk --no-cache add \
      ca-certificates \
      mailcap \
      curl

HEALTHCHECK --start-period=2s --interval=5s --timeout=3s \
  CMD curl -f http://localhost/health || exit 1

VOLUME /srv
EXPOSE 80

COPY --from=base /app/docker_config.json /.filebrowser.json
COPY --from=base /app/filebrowser /filebrowser
COPY --from=nbuild /app/dist/ ./frontend/dist/

ENTRYPOINT [ "/filebrowser" ]