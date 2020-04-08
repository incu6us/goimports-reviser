# goimports-reviser

Tool for Golang to sort goimports by 3 groups: std, general and project dependencies

Before usage:
```go
package testdata

import (
	"log"

	"github.com/incu6us/goimport-reviser/testdata/innderpkg"

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
	
	"github.com/incu6us/goimport-reviser/testdata/innderpkg"
)
```

### Use help for usage:
```bash
goimports-reviser -h
```

### Example with bash cmd:
```bash
goimports-reviser -project-name github.com/incu6us/goimport-reviser -file-path ./reviser/reviser.go 
```

### Example, to configure it with JetBrains(file watcher plugin):
![example](./images/image.png)
