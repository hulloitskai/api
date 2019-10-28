FROM alpine:3.9

# Install system dependencies.
RUN apk add --update ca-certificates tzdata

# Copy built binary.
ENV PROGRAM=server
COPY ./dist/${PROGRAM} /bin/${PROGRAM}

# Configure env and exposed ports.
ENV GOENV=production
EXPOSE 3000 6060

ENTRYPOINT $PROGRAM
