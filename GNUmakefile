GOPATH := $(shell go env | grep GOPATH | sed 's/GOPATH="\(.*\)"/\1/')
PATH := $(GOPATH)/bin:$(PATH)
export $(PATH)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

fetch: ## download makefile dependencies
	test -f $(GOPATH)/bin/goreleaser || go get -u -v github.com/goreleaser/goreleaser

clean: ## cleans previously built binaries
	rm -rf ./dist

publish: clean fetch ## publishes assets
	@if [ "${GITHUB_TOKEN}" == "" ]; then\
	  echo "GITHUB_TOKEN is not set";\
		exit 1;\
	fi
	@if [ "$(GIT_BRANCH)" != "master" ]; then\
	  echo "Current branch is: '$(GIT_BRANCH)'.  Please publish from 'master'";\
		exit 1;\
	fi
	git tag -a $(VERSION) -m "$(MESSAGE)"
	git push --follow-tags
	$(GOPATH)/bin/goreleaser

.PHONY: docs
docs: ## build docs
	rm -rf ./docs
	mkdir -p ./docs
	go run main.go doc

build: clean fetch docs ## publishes in dry run mode
	$(GOPATH)/bin/goreleaser --skip-publish --snapshot


