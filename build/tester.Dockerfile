FROM golang:1.12-alpine AS tester

## Install dependencies.
RUN apk add --update git make gcc musl-dev

## Install code checking tools.
RUN GO111MODULE=off go get \
      golang.org/x/lint/golint \
      golang.org/x/tools/cmd/goimports

ENV GO111MODULE=on GOENV=production
CMD ["make", "review", "--", "-v"]
