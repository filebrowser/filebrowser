## Multistage build: First stage fetches helper scripts
FROM alpine:3.23 AS fetcher

# download JSON.sh
RUN apk update && \
    apk --no-cache add ca-certificates && \
    wget -O /JSON.sh https://raw.githubusercontent.com/dominictarr/JSON.sh/0d5e5c77365f63809bf6e77ef44a1f34b0e05840/JSON.sh

## Second stage: Use Alpine for the final runtime environment
FROM alpine:3.23

# Install runtime dependencies. ffmpeg includes ffprobe for video thumbnails.
RUN apk --no-cache add ca-certificates ffmpeg mailcap tini-static

# Define non-root user UID and GID
ENV UID=1000
ENV GID=1000

# Create user group and user
RUN addgroup -g $GID user && \
    adduser -D -u $UID -G user user

# Copy binary, scripts, and configurations into image with proper ownership
COPY --chown=user:user filebrowser /bin/filebrowser
COPY --chown=user:user docker/common/ /
COPY --chown=user:user docker/alpine/ /
COPY --from=fetcher /JSON.sh /JSON.sh

# Create data directories, set ownership, and ensure healthcheck script is executable
RUN mkdir -p /config /database /srv && \
    chown -R user:user /config /database /srv \
    && chmod +x /healthcheck.sh

# Define healthcheck script
HEALTHCHECK --start-period=2s --interval=5s --timeout=3s CMD /healthcheck.sh

# Set the user, volumes and exposed ports
USER user

VOLUME /srv /config /database

EXPOSE 80

ENTRYPOINT [ "/sbin/tini-static", "--", "/init.sh" ]
