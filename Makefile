.PHONY: test build lint clean sample release delete-tag

VERSION ?= 

build:
	go build ./...

# Release a new version (usage: make release VERSION=v0.1.0)
release:
ifndef VERSION
	$(error VERSION is required. Usage: make release VERSION=v0.1.0)
endif
	@echo "Creating and pushing tag $(VERSION)..."
	git tag $(VERSION)
	git push origin $(VERSION)
	@echo "✓ Released $(VERSION)"

# Delete a tag locally and remotely (usage: make delete-tag VERSION=v0.1.0)
delete-tag:
ifndef VERSION
	$(error VERSION is required. Usage: make delete-tag VERSION=v0.1.0)
endif
	@echo "Deleting tag $(VERSION)..."
	git tag -d $(VERSION) || true
	git push origin --delete $(VERSION) || true
	@echo "✓ Deleted $(VERSION)"

test:
	go test -v ./...

lint:
	go vet ./...

sample:
	go run cmd/sample/main.go

clean:
	go clean
	rm -f opencdp-go
