
export PORT ?= 8000

# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## show this help
 	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

install:
	go get -u -v ./...

serve-example:
	go run examples/main.go

serve-watch-example:
	go get github.com/codegangsta/gin
	PORT=8001 gin --port ${PORT} --appPort 8001 --build ./examples
