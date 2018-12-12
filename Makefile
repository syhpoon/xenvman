BUILD=go build -ldflags "-s -w"
PKG=github.com/syhpoon/xenvman/cmd
VERSION=1.0.0

.PHONY: fmt test vet deps build prepare default toc

build:
	@$(BUILD) -o xenvman $(PKG)

toc:
	@gh-md-toc README.md

fmt:
	@mk/go-tool.sh "go fmt" Formatting

test:
	@mk/go-tool.sh "go test -vet off -cover" Testing

cover:
	@mk/cover.sh

vet:
	@mk/go-tool.sh "go vet" Vetting

prepare: fmt test vet

release:
	@env GOOS=linux GOARCH=386 $(BUILD) -o xenvman-$(VERSION)-linux-386 $(PKG)
	@env GOOS=linux GOARCH=amd64 $(BUILD) -o xenvman-$(VERSION)-linux-amd64 $(PKG)
	@env GOOS=darwin GOARCH=386 $(BUILD) -o xenvman-$(VERSION)-darwin-386 $(PKG)
	@env GOOS=darwin GOARCH=amd64 $(BUILD) -o xenvman-$(VERSION)-darwin-amd64 $(PKG)
