goimports-reviser analyzer
---

### Build
Choose one of the binary(inside `./bin` dir) to your current OS & Arch types after the Make command: 
```shell
make build-all-lint
```

### Run with `go vet`
```shell
go vet -vettool=bin/macos-amd64/goimportsreviserlint ./...
```

Output:

!['linter output'](../images/linter-example.png)
