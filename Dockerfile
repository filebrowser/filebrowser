FROM alpine:latest
RUN apk --update add ca-certificates \
                     mailcap \
                     curl \
                     git

HEALTHCHECK --start-period=2s --interval=5s --timeout=3s \
  CMD curl -f http://localhost/health || exit 1

VOLUME /srv
EXPOSE 80

COPY .docker.json /.filebrowser.json
COPY filebrowser /filebrowser

ENTRYPOINT [ "/filebrowser" ]
