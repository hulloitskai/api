FROM busybox:1.30

# Copy built binary.
COPY ./apisrv /bin/apisrv

ENTRYPOINT ["apisrv"]
