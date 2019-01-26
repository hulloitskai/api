##################################################
## BUILDER IMAGE
##################################################

FROM golang:1-alpine AS builder

ARG BINARY="api"

## Install dependencies.
RUN apk add upx gcc musl-dev git make

## Copy source files.
WORKDIR /build
COPY . .

## Install app dependencies.
ENV GO111MODULE=on
WORKDIR /build
RUN go version && make install

## Create production binary at '/build/dist/$BINARY'
RUN make build BARGS="-o ./dist/${BINARY}"

## Compress binary with UPX.
RUN upx -9 "./dist/${BINARY}"


##################################################
## PRODUCTION IMAGE
##################################################

FROM alpine:3.8 as production

ARG BINARY="api"
ARG BUILD_VERSION="unset"
ENV GOENV="production"

## Labels:
LABEL maintainer="Steven Xie <dev@stevenxie.me>"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.name="stevenxie/api"
LABEL org.label-schema.description="Personal API Server"
LABEL org.label-schema.url="https://api.stevenxie.me/"
LABEL org.label-schema.vcs-url="https://github.com/stevenxie/api"
LABEL org.label-schema.version="$BUILD_VERSION"

## Install dependencies.
RUN apk add ca-certificates

## Copy production artifacts to /api.
COPY --from=builder /build/dist/${BINARY} /usr/bin/${BINARY}

COPY ./scripts/healthcheck.sh /usr/bin/healthcheck
ENV HEALTH_ENDPOINT=http://localhost:3000
HEALTHCHECK --interval=30s --timeout=30s --start-period=10s --retries=1 \
  CMD [ "healthcheck" ]

## Expose API port.
EXPOSE 3000

## Set entrypoint.
ENV BINARY=$BINARY GOENV="production"
ENTRYPOINT "$BINARY"
