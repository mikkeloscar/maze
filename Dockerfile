FROM golang:alpine as build

RUN apk --no-cache add make git gcc musl-dev

WORKDIR /build
ADD . /build

RUN make build.linux

FROM alpine:3.12
MAINTAINER Mikkel Larsen <m@moscar.net>

RUN apk --no-cache upgrade \
    && apk --no-cache add \
        ca-certificates \
        bash \
        libarchive-tools \
        pacman

COPY store/migration/sqlite3/* /store/migration/sqlite3/

# add binary
COPY --from=build /build/build/linux/maze /

ENTRYPOINT ["/maze"]
