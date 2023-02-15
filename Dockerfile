FROM debian:bullseye-slim

# To fix go get and build with cgo
# RUN apk add --no-cache --virtual .build-deps \
#     bash \
#     gcc \
#     git \
#     musl-dev

ENV GO111MODULE=on

ADD . /app/bin
WORKDIR /app/bin


CMD ["./bin/web"]