## ----- VARIABLES -----
## Program version.
VERSION ?= latest
__GIT_DESC = git describe --tags
ifneq ($(shell $(__GIT_DESC) 2> /dev/null),)
	VERSION = $(shell $(__GIT_DESC) | cut -c 2-)
endif

## Go module name.
GOMODULE = $(shell basename "$$(pwd)")
ifeq ($(shell ls -1 go.mod 2> /dev/null),go.mod)
	GOMODULE = $(shell cat go.mod | head -1 | awk '{print $$2}')
endif

## Custom Go linker flag.
LDFLAGS = -X $(GOMODULE)/internal/info.Version=$(VERSION)

## Project variables:
GOENV        ?= development
GODEFAULTCMD =  server
DKDIR        =  ./build


## ----- TARGETS ------
## Generic:
.PHONY: default version setup install build clean run lint test review release \
        help
__ARGS = $(filter-out $@,$(MAKECMDGOALS))

default: run
version: ## Show project version (derived from 'git describe').
	@echo $(VERSION)

setup: go-setup ## Set this project up on a new environment.
	@echo "Configuring githooks..." && \
	 git config core.hooksPath .githooks && \
	 echo done
install: go-install ## Install project dependencies.

run: ## Run project (development).
	$(eval __ARGS := $(if $(__ARGS),$(__ARGS),$(GODEFAULTCMD)))
	@GOENV="$(GOENV)" $(MAKE) go-run -- $(__ARGS)
build: ## Build project.
	$(eval __ARGS := $(if $(__ARGS),$(__ARGS),$(GODEFAULTCMD)))
	@$(MAKE) go-build -- $(__ARGS)
clean: ## Clean build artifacts.
	@$(MAKE) go-clean -- $(__ARGS)

lint: go-lint ## Lint and check code.
test: ## Run tests.
	@$(MAKE) go-test -- $(__ARGS)
review: ## Lint code and run tests.
	@$(MAKE) go-review -- $(__ARGS)
release: ## Release / deploy this project.
	@echo "No release procedure defined."

## Show usage for the targets in this Makefile.
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	 awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'


## CI:
.PHONY: ci-install ci-test ci-deploy
__KB = kubectl

ci-install:
	@$(MAKE) DKENV=test dk-pull
ci-test:
	@$(MAKE) dk-test -- $(__ARGS) && \
	 $(MAKE) DKENV=ci dk-up -- --no-start && \
	 $(MAKE) DKENV=ci dk-tags
ci-deploy:
	@$(MAKE) dk-push DKENV=ci && \
	 for deploy in $(__ARGS); do \
	   $(__KB) patch deployment "$$deploy" \
	     -p "{\"spec\":{\"template\":{\"metadata\":{\"annotations\":{\"date\":\"$$(date +'%s')\"}}}}}"; \
	 done


## git-secret:
.PHONY: secrets-hide secrets-reveal
secrets-hide: ## Hides modified secret files using git-secret.
	@echo "Hiding modified secret files..." && git secret hide -m
secrets-reveal: ## Reveals secret files that were hidden using git-secret.
	@echo "Revealing hidden secret files..." && git secret reveal


## Go:
.PHONY: go-setup go-install go-deps go-build go-clean go-run go-lint go-test \
        go-bench go-review

go-setup: go-install go-deps
go-deps: ## Verify and tidy project dependencies.
	@echo "Verifying module dependencies..." && \
	 go mod verify && \
	 echo "Tidying module dependencies..." && \
	 go mod tidy && \
	 echo done
go-install:
	@echo "Downloading module dependencies..." && \
	 go mod download && \
	 echo done

GOCMDDIR     ?= ./cmd
GOBUILDDIR   ?= ./dist
GOBUILDFLAGS = -ldflags "$(LDFLAGS)"

__GOCMDNAME   = $(firstword $(__ARGS))
__GOCMD       = $(GOCMDDIR)/$(__GOCMDNAME)
__GOARGS      = $(filter-out $(__GOCMDNAME),$(__ARGS))
__GOVERIFYCMD = \
	@if [ -z $(__GOCMD) ]; then \
	  echo "No build package was specified." && exit 1; \
	fi

go-run:
	@$(__GOVERIFYCMD) && echo "Running with 'go run'..." && \
	 go run $(GOBUILDFLAGS) $(GORUNFLAGS) $(__GOCMD) $(__GOARGS)
go-build: __go-verify-cmd
	@echo "Building with 'go build'..." && \
	 go build $(GOBUILDFLAGS) -o $(GOBUILDDIR)/$(__GOCMDNAME) \
	   $(__GOCMD) $(__GOARGS) && \
	 echo done
go-clean:
	@echo "Cleaning with 'go clean'..." && \
	 go clean $(__ARGS) && \
	 echo done

go-lint:
	@if command -v goimports > /dev/null; then \
	   echo "Formatting code with 'goimports'..." && \
	   goimports -w -l . | tee /dev/stderr | xargs -0 test -z; EXIT=$$?; \
	 else \
	   echo "'goimports' not installed, skipping format step."; \
	 fi && \
	 if command -v golint > /dev/null; then \
	   echo "Linting code with 'golint'..." && \
	   golint -set_exit_status ./...; EXIT="$$((EXIT | $$?))"; \
	 else \
	   echo "'golint' not installed, skipping linting step."; \
	 fi && \
	 echo "Checking code with 'go vet'..." && go vet ./... && \
	 echo done && exit $$EXIT
go-review:
	@$(MAKE) go-lint && $(MAKE) go-test -- $(__ARGS)

GOCOVERFILE   ?= coverage.out
GOTESTTIMEOUT ?= 20s
GOTESTFLAGS   ?= -race

__GOTEST = \
  go test \
	  -coverprofile="$(GOCOVERFILE)" -covermode=atomic \
    -timeout="$(GOTESTTIMEOUT)" \
	  $(GOBUILDFLAGS) $(GOTESTFLAGS)
go-test:
	@echo "Running tests with 'go test':" && $(__GOTEST) ./... $(__ARGS)
go-bench: ## Run benchmarks.
	@echo "Running benchmarks with 'go test -bench=.'..." && \
	 $(__GOTEST) -run=^$$ -bench=. -benchmem ./... $(__ARGS)


## Docker:
.PHONY: dk-pull dk-push dk-build dk-build-push dk-clean dk-tags dk-up \
        dk-build-up dk-down dk-logs dk-test

DKDIR ?= .

__DKFILE = $(DKDIR)/docker-compose.yml
ifeq ($(DKENV),test)
	__DKFILE = $(DKDIR)/docker-compose.test.yml
endif
ifeq ($(DKENV),ci)
	__DKFILE = $(DKDIR)/docker-compose.build.yml
endif

__DK        = docker
__DKCMP     = docker-compose -f "$(__DKFILE)"
__DKCMP_VER = VERSION="$(VERSION)" $(__DKCMP)
__DKCMP_LST = VERSION=latest $(__DKCMP)

dk-pull: ## Pull latest Docker images from registry.
	@echo "Pulling latest images from registry..." && \
	 $(__DKCMP_LST) pull $(__ARGS)
dk-push: ## Push new Docker images to registry.
	@if git describe --exact-match --tags > /dev/null 2>&1; then \
	   echo "Pushing versioned images to registry (:$(VERSION))..." && \
	   $(__DKCMP_VER) push $(__ARGS); \
	 fi && \
	 echo "Pushing latest images to registry (:latest)..." && \
	 $(__DKCMP_LST) push $(__ARGS) && \
	 echo done

dk-build: ## Build and tag Docker images.
	@echo "Building images..." && \
	 $(__DKCMP_VER) build --parallel --compress $(__ARGS) && \
	 echo done && \
	 $(MAKE) dk-tags
dk-clean: ## Clean up unused Docker data.
	@echo "Cleaning unused data..." && $(__DK) system prune $(__ARGS)
dk-tags: ## Tag versioned Docker images with ':latest'.
	@echo "Tagging versioned images with ':latest'..." && \
	 images="$$($(__DKCMP_VER) config | egrep image | awk '{print $$2}')" && \
	 for image in $$images; do \
	   if [ -z "$$($(__DK) images -q "$$image" 2> /dev/null)" ]; then \
	     continue; \
	   fi && \
	   echo "$$image" | sed -e 's/:.*$$/:latest/' | \
	     xargs $(__DK) tag "$$image"; \
	 done && \
	 echo done
dk-build-push: dk-build dk-push ## Build and push new Docker images.

__DK_UP = $(__DKCMP_VER) up
dk-up: ## Start up containerized services.
	@echo "Bringing up services..." && $(__DK_UP) $(__ARGS) && echo done
dk-down: ## Shut down containerized services.
	@echo "Bringing down services..." && \
	 $(__DKCMP_VER) down $(__ARGS) && \
	 echo done
dk-build-up: ## Build new images, then start them.
	@echo "Building and bringing up services..." && \
	 $(__DK_UP) --build $(__ARGS) && \
	 echo done

dk-logs: ## Show logs for containerized services.
	@$(__DKCMP_VER) logs -f $(__ARGS)
dk-test: ## Test using 'docker-compose.test.yml'.
	$(eval __DKFILE = $(DKDIR)/docker-compose.test.yml)
	@if [ -s "$(__DKFILE)" ]; then \
	   echo "Running containerized service tests..." && \
	   for svc in $$($(__DKCMP_LST) config --services); do \
	     if ! $(__DK_UP) --abort-on-container-exit $(__ARGS) "$$svc"; \
	       then exit -1; \
	     fi \
	   done; \
	 fi && \
	 echo done


## HACKS:
## These targets are hacks that allow for Make targets to receive
## pseudo-arguments.
.PHONY: __FORCE
__FORCE:
%: __FORCE
	@:
