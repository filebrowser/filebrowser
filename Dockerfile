FROM ubuntu:latest
RUN apt update
RUN apt install -y ca-certificates sendmail curl

HEALTHCHECK --start-period=2s --interval=5s --timeout=3s \
  CMD curl -f http://localhost:9091/health || exit 1

VOLUME /srv

COPY .docker.json /.filebrowser.json
COPY filebrowser /filebrowser
COPY run.sh /run.sh

ENTRYPOINT [ "/run.sh" ]
