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

.PHONY: build-all-lint
build-all-lint: build-lint-windows-386 build-lint-windows-amd64 build-lint-macos-amd64 build-lint-macos-arm64 build-lint-linux-386 build-lint-linux-amd64 build-lint-linux-arm64

.PHONY: build-lint-windows-386
build-lint-windows-386:
	GOOS=windows GOARCH=386 go build -o bin/windows-386/goimportsreviserlint.exe ./pkg/goanalysis/analyzer.go

.PHONY: build-lint-windows-amd64
build-lint-windows-amd64:
	GOOS=windows GOARCH=amd64 go build -o bin/windows-amd64/goimportsreviserlint.exe ./pkg/goanalysis/analyzer.go

.PHONY: build-lint-macos-amd64
build-lint-macos-amd64:
	GOOS=darwin GOARCH=amd64 go build -o bin/macos-amd64/goimportsreviserlint ./pkg/goanalysis/analyzer.go

.PHONY: build-lint-macos-arm64
build-lint-macos-arm64:
	GOOS=darwin GOARCH=arm64 go build -o bin/macos-arm64/goimportsreviserlint ./pkg/goanalysis/analyzer.go

.PHONY: build-lint-linux-386
build-lint-linux-386:
	GOOS=linux GOARCH=386 go build -o bin/linux-386/goimportsreviserlint ./pkg/goanalysis/analyzer.go

.PHONY: build-lint-linux-amd64
build-lint-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o bin/linux-amd64/goimportsreviserlint ./pkg/goanalysis/analyzer.go

.PHONY: build-lint-linux-arm64
build-lint-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -o bin/linux-arm64/goimportsreviserlint ./pkg/goanalysis/analyzer.go

.PHONY: build-macos-amd64
build-macos-amd64:
	GOOS=darwin GOARCH=amd64 go build -o bin/macos-amd64/goimports-reviser .

.PHONY: update-std-package-list
update-std-package-list:
	@go run -tags gen github.com/incu6us/goimports-reviser/v3/pkg/std/gen
