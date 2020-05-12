# goimports-reviser
!['Status Badge'](https://github.com/incu6us/goimports-reviser/workflows/build/badge.svg)
!['Release Badge'](https://github.com/incu6us/goimports-reviser/workflows/release/badge.svg)
!['Quality Badge'](https://goreportcard.com/badge/github.com/incu6us/goimports-reviser)

Tool for Golang to sort goimports by 3 groups: std, general and project dependencies.
Also formatting for your code will be prepared(so, you don't need to use `gofmt` or `goimports` separately). 

# Install
```bash
$ brew tap incu6us/homebrew-tap
$ brew install incu6us/homebrew-tap/goimports-reviser
```


Before usage:
```go
package testdata

import (
	"log"

	"github.com/incu6us/goimports-reviser/testdata/innderpkg"

	"bytes"

	"github.com/pkg/errors"
)
``` 

After usage:
```go
package testdata

import (
	"bytes"
	"log"
	
	"github.com/pkg/errors"
	
	"github.com/incu6us/goimports-reviser/testdata/innderpkg"
)
```

### Use help for details:
```bash
goimports-reviser -h
```

### Example with bash cmd:
```bash
goimports-reviser -project-name github.com/incu6us/goimports-reviser -file-path ./reviser/reviser.go 
```

### Example, to configure it with JetBrains IDEs (via file watcher plugin):
![example](./images/image.png)
