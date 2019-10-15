# == Variables ==
# Program version.
__TAG = $(shell git describe --tags --always 2> /dev/null)
ifneq ($(__TAG),)
	VERSION ?= $(shell echo "$(__TAG)" | cut -c 2-)
else
	VERSION ?= undefined
endif

# Go module name.
GOMODULE = $(shell basename "$$(pwd)")
ifeq ($(shell ls -1 go.mod 2> /dev/null),go.mod)
	GOMODULE = $(shell cat go.mod | head -1 | awk '{print $$2}')
endif

# Custom Go linker flag.
LDFLAGS = -X $(GOMODULE)/internal.Version=$(VERSION)

# Project variables:
GOENV        ?= development
GODEFAULTCMD =  server


# == Targets ==
# Generic:
.PHONY: __default __unknown setup install build clean run lint test review \
        help version
__ARGS = $(filter-out $@,$(MAKECMDGOALS))

__default:
	@$(MAKE) lint -- $(__ARGS)
__unknown:
	@echo "Target '$(__ARGS)' not configured."

setup: go-setup # Set this project up on a new environment.
	@echo "Configuring githooks..." && \
	 git config core.hooksPath .githooks && \
	 echo done
install: # Install project dependencies.
	@$(MAKE) go-install -- $(__ARGS) && \
	 $(MAKE) go-generate

run: # Run project (development).
	$(eval __ARGS := $(if $(__ARGS),$(__ARGS),$(GODEFAULTCMD)))
	@GOENV="$(GOENV)" $(MAKE) go-run -- $(__ARGS)
build: # Build project.
	$(eval __ARGS := $(if $(__ARGS),$(__ARGS),$(GODEFAULTCMD)))
	@$(MAKE) go-build -- $(__ARGS)
clean: # Clean build artifacts.
	@$(MAKE) go-clean -- $(__ARGS)

lint: # Lint and check code.
	@$(MAKE) go-lint -- $(__ARGS) && \
	 $(MAKE) proto-lint -- $(__ARGS)
test: # Run tests.
	@$(MAKE) go-test -- $(__ARGS)
review: # Lint code and run tests.
	@$(MAKE) go-review -- $(__ARGS)

# Show usage for the targets in this Makefile.
help:
	@grep -E '^[a-zA-Z_-]+:.*?# .*$$' $(MAKEFILE_LIST) | \
	 awk 'BEGIN {FS = ":.*?# "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
version: # Show project version (derived from 'git describe').
	@echo $(VERSION)


# git-secret:
.PHONY: secrets-hide secrets-reveal
secrets-hide: # Hides modified secret files using git-secret.
	@echo "Hiding modified secret files..." && git secret hide -m $(__ARGS)
secrets-reveal: # Reveals secret files that were hidden using git-secret.
	@echo "Revealing hidden secret files..." && git secret reveal $(__ARGS)


# Go:
.PHONY: go-setup go-deps go-install go-generate go-build go-clean go-run \
        go-lint go-test go-bench go-review

# Export environment variables to configure the Go toolchain.
__GOENV = if [ -n "$(GOPRIVATE)" ]; then export GOPRIVATE="$(GOPRIVATE)"; fi

go-shell: SHELL := /usr/bin/env bash
go-shell: # Launch a shell with a env preset for the Go toolchain.
	@$(__GOENV) && \
	 bash --rcfile \
	   <(echo '. $$HOME/.bashrc; PS1="\[\e[37m\]go:\[\e[39m\] $$PS1"') && \
	 exit $$?

go-setup: go-install go-deps
go-deps: # Verify and tidy project dependencies.
	@$(__GOENV) && \
	 echo "Verifying Go module dependencies..." && \
	 go mod verify && \
	 echo "Tidying Go module dependencies..." && \
	 go mod tidy && \
	 echo done
go-install:
	@$(__GOENV) && \
	 echo "Downloading Go module dependencies..." && \
	 go mod download && \
	 echo done
go-generate: # Generate Go source files.
	@echo "Generating Go files..." && \
	 go generate $(__ARGS) ./... && \
	 echo done

GOCMDDIR     ?= ./cmd
GOBUILDDIR   ?= ./dist
GOBUILDFLAGS  = -trimpath -ldflags "$(LDFLAGS)"

__GOCMDNAME   = $(firstword $(__ARGS))
__GOCMD       = $(GOCMDDIR)/$(__GOCMDNAME)
__GOARGS      = $(filter-out $(__GOCMDNAME),$(__ARGS))
__GOVERIFYCMD = \
  if [ -z $(__GOCMD) ]; then \
    echo "No build package was specified." && exit 1; \
  fi

go-run:
	@$(__GOENV) && $(__GOVERIFYCMD) && \
	 echo "Running with 'go run'..." && \
	 go run $(GOBUILDFLAGS) $(GORUNFLAGS) $(__GOCMD) $(__GOARGS)
go-build:
	@$(__GOENV) && \
	 echo "Building with 'go build'..." && \
	 go build $(GOBUILDFLAGS) -o $(GOBUILDDIR)/$(__GOCMDNAME) \
	   $(__GOCMD) $(__GOARGS) && \
	 echo done
go-clean:
	@$(__GOENV) && \
	echo "Cleaning with 'go clean'..." && \
	 go clean $(__ARGS) && \
	 echo done

go-lint:
	@$(__GOENV) && \
	 if command -v goimports > /dev/null; then \
	   echo "Formatting Go code with 'goimports'..." && \
	   goimports -w -l $$(find . -name '*.go' | grep -v '.pb.go') \
	     | tee /dev/fd/2 \
	     | xargs -0 test -z; EXIT=$$?; \
	 else \
	   echo "'goimports' not installed, skipping format step."; \
	 fi && \
	 if command -v revive > /dev/null; then \
	   echo "Linting Go code with 'revive'..." && \
	   revive -config .revive.toml ./...; EXIT="$$((EXIT | $$?))"; \
	 elif command -v golint > /dev/null; then \
	   echo "Linting Go code with 'golint'..." && \
	   golint -set_exit_status ./...; EXIT="$$((EXIT | $$?))"; \
	 else \
	   echo "Neither 'revive' nor 'golint' is installed, skipping linting step."; \
	 fi && \
	 echo "Checking code with 'go vet'..." && go vet ./... && \
	 echo done && exit $$EXIT
go-review:
	@$(MAKE) go-lint && $(MAKE) go-test -- $(__ARGS)

GOTESTTIMEOUT ?= 20s
GOTESTFLAGS   ?= -race

__GOTEST = \
  go test \
    -covermode=atomic \
    -timeout="$(GOTESTTIMEOUT)" \
    $(GOBUILDFLAGS) $(GOTESTFLAGS)
go-test:
	@$(__GOENV) && \
	 echo "Running tests with 'go test':" && \
	 $(__GOTEST) ./... $(__ARGS)
go-bench: # Run benchmarks.
	@$(__GOENV) && \
	 echo "Running benchmarks with 'go test -bench=.'..." && \
	 $(__GOTEST) -run=^$$ -bench=. -benchmem ./... $(__ARGS)


# Protobuf:
.PHONY: proto-lint
__PROTOTOOL = prototool

proto-lint:
	@echo "Formatting proto3 files with 'prototool'..." && \
	 $(__PROTOTOOL) format -l -- $(__ARGS); EXIT=$$?; \
	 $(__PROTOTOOL) format -w -- $(__ARGS) && \
	 echo "Linting proto3 files with 'prototool'..." && \
	 $(__PROTOTOOL) lint -- $(__ARGS); EXIT="$$((EXIT | $$?))"; \
	 echo done && exit $$EXIT


# HACKS:
# These targets are hacks that allow for Make targets to receive
# pseudo-arguments.
.PHONY: __FORCE
__FORCE:
%: __FORCE
	@:
