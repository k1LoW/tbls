FROM alpine:latest

ENTRYPOINT ["tbls"]
WORKDIR /work
VOLUME ["/work"]

ARG DOCKER_TAG
ENV TBLS_VERION=v$DOCKER_TAG

RUN apk add bash curl

SHELL ["/bin/bash", "-c"]

RUN set -x \
        && source <(curl -sL https://git.io/use-tbls) \
        && which tbls | xargs -I{} mv {} /usr/local/bin/tbls \
        && apk del bash curl

SHELL ["/bin/sh", "-c"]
