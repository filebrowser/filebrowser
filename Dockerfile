FROM alpine:latest as builder
RUN apk --update add ca-certificates \
                     mailcap \
                     curl \
                     jq \
                     libc6-compat \
                     make \
                     nodejs \
                     npm \
                     bash \
                     ncurses \
                     go \
                     git
WORKDIR /build

COPY ./ /build

RUN go mod download

RUN make build

VOLUME /srv
EXPOSE 80

FROM alpine:latest as target

WORKDIR /app

COPY healthcheck.sh /healthcheck.sh
RUN chmod +x /healthcheck.sh  # Make the script executable

HEALTHCHECK --start-period=2s --interval=5s --timeout=3s \
    CMD /healthcheck.sh || exit 1

COPY docker_config.json .filebrowser.json
COPY --from=builder /build/filebrowser filebrowser

ENTRYPOINT [ "./filebrowser" ]