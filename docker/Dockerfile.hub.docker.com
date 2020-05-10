FROM alpine:latest

ENTRYPOINT ["tbls"]
WORKDIR /work
VOLUME ["/work"]

ARG DOCKER_TAG

RUN apk add bash curl

SHELL ["/bin/bash", "-c"]

RUN set -x \
        && curl -L https://git.io/dpkg-i-from-url | bash -s -- https://github.com/k1LoW/tbls/releases/download/v$DOCKER_TAG/tbls_$v$DOCKER_TAG-1_amd64.deb \
        && apk del bash curl

SHELL ["/bin/sh", "-c"]
