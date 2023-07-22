BINARY_NAME=ping
VERSION=1.0.0
GOARCH=amd64

default:
	go build -o $(BINARY_NAME) -v

test:
	go test -v

linux:
	GOOS=linux GOARCH=$(GOARCH) go build -o build/$(BINARY_NAME)-$(VERSION)-linux-$(GOARCH) -v

windows:
	GOOS=windows GOARCH=$(GOARCH) go build -o build/$(BINARY_NAME)-$(VERSION)-windows-$(GOARCH).exe -v

darwin:
	GOOS=darwin GOARCH=$(GOARCH) go build -o build/$(BINARY_NAME)-$(VERSION)-darwin-$(GOARCH) -v

clean:
	go clean
	rm -rf build/

release: linux windows darwin
	git tag $(VERSION)
	git push origin $(VERSION)
