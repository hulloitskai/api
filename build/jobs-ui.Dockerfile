##################################################
## BUILDER IMAGE
##################################################

FROM golang:1-alpine AS builder

ARG BINARY="workwebui"

## Install dependencies.
RUN apk add upx gcc musl-dev git

## Install app dependencies.
RUN go version && \
    go get -u github.com/gocraft/work/cmd/workwebui && ls -l && \
    mkdir -p /build/dist && \
    mv "./bin/${BINARY}" "/build/dist/${BINARY}"

## Compress binary with UPX.
RUN upx -9 "/build/dist/${BINARY}"


##################################################
## PRODUCTION IMAGE
##################################################

FROM alpine:3.8 as production

ARG BINARY="workwebui"
ARG BUILD_VERSION="unset"
ENV GOENV="production"

## Labels:
LABEL maintainer="Steven Xie <dev@stevenxie.me>"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.name="stevenxie/api-jobs-ui"
LABEL org.label-schema.description="API Jobserver UI"
LABEL org.label-schema.url="https://api.stevenxie.me/"
LABEL org.label-schema.vcs-url="https://github.com/stevenxie/api"
LABEL org.label-schema.version="$BUILD_VERSION"

## Install dependencies.
RUN apk add ca-certificates

## Copy production artifacts to /api.
COPY --from=builder /build/dist/${BINARY} /usr/bin/${BINARY}

COPY ./scripts/healthcheck.sh /usr/bin/healthcheck
ENV HEALTH_ENDPOINT=http://localhost:80
HEALTHCHECK --interval=30s --timeout=30s --start-period=10s --retries=1 \
  CMD [ "healthcheck" ]

## Expose API port.
EXPOSE 80

## Set entrypoint.
ENV BINARY="$BINARY" GOENV="production" NAMESPACE="api"
CMD workwebui -redis="${REDIS_ADDR}" -ns="${NAMESPACE}" -listen=":80"
