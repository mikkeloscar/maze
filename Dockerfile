FROM alpine:3.5
MAINTAINER Mikkel Larsen <m@moscar.net>

RUN apk --no-cache upgrade \
    && apk --no-cache add \
        ca-certificates \
        bash \
        libarchive-tools \
        pacman

COPY store/migration/sqlite3/* /store/migration/sqlite3/

# add binary
ADD build/linux/maze /

ENTRYPOINT ["/maze"]
