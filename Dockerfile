FROM golang:1.12-alpine as builder

# To fix go get and build with cgo
RUN apk add --no-cache --virtual .build-deps \
    bash \
    gcc \
    git \
    musl-dev

COPY . /app
WORKDIR /app

EXPOSE 5000

# CMD ["go", "run", "./app/cmd/web"]