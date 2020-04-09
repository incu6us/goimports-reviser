.PHONY: go-generate
go-generate:
	@go generate -tags gen ./...

.PHONY: go-test
go-test:
	@go test -race -v -cover ./...
