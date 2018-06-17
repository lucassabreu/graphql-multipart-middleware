all: help

export PORT ?= 8000

# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## show this help
 	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

install: ## install package dependencies
	go get -u -v -t ./...

serve-example: ## start the example server
	go run examples/main.go

serve-watch-example: ## start the example server watching for changes
	go get github.com/codegangsta/gin
	PORT=8001 gin --port ${PORT} --appPort 8001 --build ./examples

tests: ## run the package's tests
	go test -v -race .

coverage: ## calcs the coverage for the package
	go get golang.org/x/tools/cmd/cover
	go get github.com/mattn/goveralls
	go test -v -covermode=count -coverprofile=coverage.out

send-statistics: ## send statistics
	goveralls -coverprofile=coverage.out -service=travis-ci -repotoken ${COVERALLS_TOKEN}

