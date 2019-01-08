FROM scratch

COPY --from=filebrowser/dev /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

VOLUME /tmp
VOLUME /srv
EXPOSE 80

COPY .filebrowser.docker.json /.filebrowser.json
COPY filebrowser /filebrowser

ENTRYPOINT [ "/filebrowser" ]
