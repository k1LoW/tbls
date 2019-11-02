FROM alpine:latest

ENTRYPOINT ["tbls"]
WORKDIR /work
VOLUME ["/work"]

ENV TBLS_VERSION=v1.20.0
RUN set -x \
        && wget https://github.com/k1LoW/tbls/releases/download/${TBLS_VERSION}/tbls_${TBLS_VERSION}_linux_amd64.tar.gz \
        && tar xzf tbls_${TBLS_VERSION}_linux_amd64.tar.gz \
        && mv tbls /usr/local/bin/ \
        && rm tbls_${TBLS_VERSION}_linux_amd64.tar.gz
