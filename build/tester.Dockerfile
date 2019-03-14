FROM golang:1 AS tester

## Install code checking tools.
RUN GO111MODULE=off go get \
      golang.org/x/lint/golint \
      golang.org/x/tools/cmd/goimports

ENV GO111MODULE=on GOENV=production
CMD ["make", "review", "TARGS=-v -race"]
