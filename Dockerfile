FROM alpine:latest
RUN apk --update add ca-certificates
RUN apk --update add mailcap

VOLUME /srv
EXPOSE 80

COPY filebrowser /filebrowser

# Create appuser.
ENV USER=root
ENV GROUP=root
ENV UID=0
ENV GID=0
ENV UMASK=022

RUN if [ "$GID" -ne 0 ]; then \
    addgroup \
    -g "${GID}" \
    "${GROUP}" ; \
    fi;

RUN adduser \
    -g "" \
    -D \
    -G "${GROUP}" \
    -H \
    -h "/nonexistent" \
    -s "/sbin/nologin" \
    -u "${UID}" \
    "${USER}"

USER ${USER}:${GROUP}

RUN umask ${UMASK}

ENTRYPOINT [ "/filebrowser" ]