FROM alpine:latest

RUN apk add --no-cache bash curl git

ENTRYPOINT ["tbls"]
CMD [ "-h" ]

COPY tbls_*.apk /tmp/
RUN apk add --allow-untrusted /tmp/tbls_*.apk
