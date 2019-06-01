FROM alpine:3.9

# Install system dependencies.
RUN apk add --update ca-certificates tzdata

# Copy built binary.
COPY ./dist/apisrv /bin/apisrv

ENTRYPOINT ["apisrv"]
