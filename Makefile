.PHONY: default
default:
	@echo Commands:
	@echo "make fmt - Format Go code"
	@echo "make test Run unit tests"
	@echo "make vet - Run go vet"
	@echo "make deps" - Update vendor dependencies
	@echo "make build" - Build xenvman


.PHONY: fmt test vet deps build prepare

fmt:
	@mk/go-tool.sh "go fmt" Formatting

test:
	@mk/go-tool.sh "go test -vet off -cover" Testing

vet:
	@mk/go-tool.sh "go vet" Vetting

deps:
	git submodule update --init --recursive

prepare: fmt test vet

build: deps
	@go build -ldflags "-s -w" -o xenvman github.com/syhpoon/xenvman/cmd
