FROM alpine:3.22

RUN apk update && \
  apk --no-cache add ca-certificates mailcap curl jq tini

# Make user and create necessary directories
ENV UID=1000
ENV GID=1000

RUN addgroup -g $GID user && \
  adduser -D -u $UID -G user user && \
  mkdir -p /config /database /srv && \
  chown -R user:user /config /database /srv

# Copy files and set permissions
COPY filebrowser /bin/filebrowser
COPY docker/common/ /
COPY docker/alpine/ /

RUN chown -R user:user /bin/filebrowser /defaults healthcheck.sh init.sh

# Define healthcheck script
HEALTHCHECK --start-period=2s --interval=5s --timeout=3s CMD /healthcheck.sh

# Set the user, volumes and exposed ports
USER user

VOLUME /srv /config /database

EXPOSE 80

ENTRYPOINT [ "tini", "--", "/init.sh", "filebrowser", "--config", "/config/settings.json" ]
