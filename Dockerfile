FROM node:alpine as frontend

WORKDIR /filebrowser/frontend
COPY frontend/ .

# TODO: Remove when frontend dependencies are updated
ENV NODE_OPTIONS --openssl-legacy-provider

RUN npm install && \
    npx browserslist@latest --update-db &&\
    npm run build


FROM golang:alpine as backend

WORKDIR /filebrowser
COPY . .
COPY --from=frontend /filebrowser /filebrowser
RUN go mod download && \
    go build


FROM alpine:latest
RUN apk --update add ca-certificates \
                     mailcap \
                     curl \
                     jq

COPY healthcheck.sh /healthcheck.sh
RUN chmod +x /healthcheck.sh  # Make the script executable

HEALTHCHECK --start-period=2s --interval=5s --timeout=3s \
    CMD /healthcheck.sh || exit 1

VOLUME /srv
EXPOSE 80

COPY docker_config.json /.filebrowser.json
COPY --from=backend /filebrowser/filebrowser /filebrowser

ENTRYPOINT [ "/filebrowser" ]