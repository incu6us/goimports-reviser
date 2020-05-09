.PHONY: go-generate
go-generate:
	@go generate -tags gen ./...

.PHONY: go-test
go-test:
	@go test -race -v -cover ./...

# Create dist only locally
.PHONY: release-check
release-check:
	@goreleaser --snapshot --skip-publish --rm-dist

# Run without without publishing
.PHONY: release-dry-run
release-dry-run:
	@goreleaser release --skip-publish --rm-dist
