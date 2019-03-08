##################################################
## BUILDER IMAGE
##################################################

FROM golang:1-alpine AS builder

ARG BINARY="jobserver"

## Install dependencies.
RUN apk add --update upx gcc musl-dev git make

## Copy source files.
WORKDIR /build
COPY . .

## Install app dependencies.
ENV GO111MODULE=on
WORKDIR /build
RUN go version && make install

## Create production binary at '/build/dist/$BINARY'
RUN make build BDIR="./cmd/jobserver" BARGS="-o ./dist/${BINARY}"

## Compress binary with UPX.
RUN upx -9 "./dist/${BINARY}"


##################################################
## PRODUCTION IMAGE
##################################################

FROM alpine:3.8 as production

ARG BINARY="jobserver"
ARG BUILD_VERSION="unset"
ENV GOENV="production"

## Labels:
LABEL maintainer="Steven Xie <dev@stevenxie.me>"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.name="stevenxie/api-jobs"
LABEL org.label-schema.url="https://api.stevenxie.me/"
LABEL org.label-schema.vcs-url="https://github.com/stevenxie/api"
LABEL org.label-schema.version="$BUILD_VERSION"

## Install dependencies.
RUN apk add --no-cache ca-certificates

## Copy production artifacts to /usr/bin/.
COPY --from=builder /build/dist/${BINARY} /usr/bin/${BINARY}

## Expose API port.
EXPOSE 3000

## Set entrypoint.
ENV BINARY=$BINARY GOENV="production"
ENTRYPOINT "$BINARY"
