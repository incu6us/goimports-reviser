<br/>
<div align="center">
  <a href="https://vshymanskyy.github.io/StandWithUkraine">
    <img src="images/dove.png" class="this" alt="Stand With Ukraine" style="width: 60%;">
  </a>
  <h3 align="center"><a href="https://vshymanskyy.github.io/StandWithUkraine">🇺🇦 #StandWithUkraine 🇺🇦</a></h3>
</div>

---

# goimports-reviser [![Tweet](https://img.shields.io/twitter/url/http/shields.io.svg?style=social)](https://twitter.com/intent/tweet?text=Right%20golang%20imports%20sorting%20and%20code%20formatting%20tool%20(goimports%20alternative)&url=https://github.com/incu6us/goimports-reviser&hashtags=golang,code,goimports-reviser,goimports,gofmt,developers)
[![#StandWithUkraine](https://raw.githubusercontent.com/vshymanskyy/StandWithUkraine/main/badges/StandWithUkraine.svg)](https://vshymanskyy.github.io/StandWithUkraine)
!['Status Badge'](https://github.com/incu6us/goimports-reviser/workflows/build/badge.svg)
!['Release Badge'](https://github.com/incu6us/goimports-reviser/workflows/release/badge.svg)
!['Quality Badge'](https://goreportcard.com/badge/github.com/incu6us/goimports-reviser)
[![codecov](https://codecov.io/gh/incu6us/goimports-reviser/branch/master/graph/badge.svg)](https://codecov.io/gh/incu6us/goimports-reviser)
![GitHub All Releases](https://img.shields.io/github/downloads/incu6us/goimports-reviser/total?color=green)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/incu6us/goimports-reviser?color=green)
[![goimports-reviser](https://snapcraft.io//goimports-reviser/badge.svg)](https://snapcraft.io/goimports-reviser)
![license](https://img.shields.io/github/license/incu6us/goimports-reviser)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go) 

<a href="https://www.buymeacoffee.com/slavka" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/v2/default-yellow.png" alt="Buy Me A Coffee" style="height: 60px !important;width: 217px !important;"></a>

!['logo'](./images/reviser-muscot_200.png)


Tool for Golang to sort goimports by 3-4 groups(with own [linter](linter/README.md)): std, general, company(which is optional) and project dependencies.
Also, formatting for your code will be prepared(so, you don't need to use `gofmt` or `goimports` separately). 
Use additional options `-rm-unused` to remove unused imports and `-set-alias` to rewrite import aliases for versioned packages or for packages with additional prefix/suffix(example: `opentracing "github.com/opentracing/opentracing-go"`).
`-company-prefixes` - will create group for company imports(libs inside your organization). Values should be comma-separated.


## Configuration:
### Cmd
```bash
goimports-reviser -rm-unused -set-alias -format ./reviser/reviser.go
```

You can also apply rules to a dir or recursively apply using ./... as a target:
```bash
goimports-reviser -rm-unused -set-alias -format -recursive reviser
```

```bash
goimports-reviser -rm-unused -set-alias -format ./...
```

You can also apply rules to multiple targets:
```bash
goimports-reviser -rm-unused -set-alias -format ./reviser/reviser.go ./pkg/...
```

### Example, to configure it with JetBrains IDEs (via file watcher plugin):
![example](./images/image.png)


### Options:
```text
Usage of goimports-reviser:
  -apply-to-generated-files
    	Apply imports sorting and formatting(if the option is set) to generated files. Generated file is a file with first comment which starts with comment '// Code generated'. Optional parameter.
  -company-prefixes string
    	Company package prefixes which will be placed after 3rd-party group by default(if defined). Values should be comma-separated. Optional parameters.
  -excludes string
    	Exclude files or dirs, example: '.git/,proto/*.go'.
  -file-path string
    	Deprecated. Put file name as an argument(last item) of command line.
  -format
    	Option will perform additional formatting. Optional parameter.
  -imports-order string
    	Your imports groups can be sorted in your way.
    	std - std import group;
    	general - libs for general purpose;
    	company - inter-org or your company libs(if you set '-company-prefixes'-option, then 4th group will be split separately. In other case, it will be the part of general purpose libs);
    	project - your local project dependencies;
    	blanked - imports with "_" alias;
    	dotted - imports with "." alias.
    	Optional parameter. (default "std,general,company,project")
  -list-diff
    	Option will list files whose formatting differs from goimports-reviser. Optional parameter.
  -local string
    	Deprecated
  -output string
    	Can be "file", "write" or "stdout". Whether to write the formatted content back to the file or to stdout. When "write" together with "-list-diff" will list the file name and write back to the file. Optional parameter. (default "file")
  -project-name string
    	Your project name(ex.: github.com/incu6us/goimports-reviser). Optional parameter.
  -recursive
    	Apply rules recursively if target is a directory. In case of ./... execution will be recursively applied by default. Optional parameter.
  -rm-unused
    	Remove unused imports. Optional parameter.
  -separate-named
        Separate named imports from their group with a new line. Optional parameter.
  -set-alias
    	Set alias for versioned package names, like 'github.com/go-pg/pg/v9'. In this case import will be set as 'pg "github.com/go-pg/pg/v9"'. Optional parameter.
  -set-exit-status
    	set the exit status to 1 if a change is needed/made. Optional parameter.
  -use-cache
    	Use cache to improve performance. Optional parameter.
  -version
    	Show version.
```

## Install
### With Go
```bash
go install -v github.com/incu6us/goimports-reviser/v3@latest
```

### With Brew
```bash
brew tap incu6us/homebrew-tap
brew install incu6us/homebrew-tap/goimports-reviser
```

### With Snap
```bash
snap install goimports-reviser
```

## Examples
Before usage:
```go
package testdata

import (
	"log"

	"github.com/incu6us/goimports-reviser/testdata/innderpkg"

	"bytes"

	"golang.org/x/exp/slices"
)
``` 

After usage:
```go
package testdata

import (
	"bytes"
	"log"

	"golang.org/x/exp/slices"

	"github.com/incu6us/goimports-reviser/testdata/innderpkg"
)
```

Comments(not Docs) for imports is acceptable. Example:
```go
package testdata

import (
    "fmt" // comments to the package here
)
```  

### Example with `-company-prefixes`-option

Before usage:

```go
package testdata // goimports-reviser/testdata

import (
	"fmt" //fmt package
	"golang.org/x/exp/slices" //custom package
	"github.com/incu6us/goimports-reviser/pkg" // this is a company package which is not a part of the project, but is a part of your organization
	"goimports-reviser/pkg"
)
```

After usage:
```go
package testdata // goimports-reviser/testdata

import (
	"fmt" // fmt package

	"golang.org/x/exp/slices" // custom package

	"github.com/incu6us/goimports-reviser/pkg" // this is a company package which is not a part of the project, but is a part of your organization

	"goimports-reviser/pkg"
)
```

### Example with `-imports-order std,general,company,project,blanked,dotted`-option

Before usage:

```go
package testdata // goimports-reviser/testdata

import (
	_ "github.com/pkg1"
	. "github.com/pkg2"
	"fmt" //fmt package
	"golang.org/x/exp/slices" //custom package
	"github.com/incu6us/goimports-reviser/pkg" // this is a company package which is not a part of the project, but is a part of your organization
	"goimports-reviser/pkg"
)
```

After usage:
```go
package testdata // goimports-reviser/testdata

import (
	"fmt" // fmt package

	"golang.org/x/exp/slices" // custom package

	"github.com/incu6us/goimports-reviser/pkg" // this is a company package which is not a part of the project, but is a part of your organization

	"goimports-reviser/pkg"

	_ "github.com/pkg1"

	. "github.com/pkg2"
)
```

### Example with `-format`-option

Before usage:
```go
package main
func test(){
}
func additionalTest(){
}
```

After usage:
```go
package main

func test(){
}

func additionalTest(){
}
```

### Example with `-separate-named`-option

Before usage:

```go
package testdata // goimports-reviser/testdata

import (
	"fmt"
	"github.com/incu6us/goimports-reviser/pkg"
	extpkg "google.com/golang/pkg"
	"golang.org/x/exp/slices"
	extslice "github.com/PeterRK/slices"
)
```

After usage:
```go
package testdata // goimports-reviser/testdata

import (
	"fmt"

	"github.com/incu6us/goimports-reviser/pkg"
	"golang.org/x/exp/slices"

	extpkg "google.com/golang/pkg"
	extslice "github.com/PeterRK/slices"
)
```
---
## Contributors

A big thank you to all the amazing people who contributed!

<a href="https://github.com/incu6us/goimports-reviser/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=incu6us/goimports-reviser" />
</a>

## Give a Star! ⭐
If you like or are using this project, please give it a **star**.


### Stargazers

[![Stargazers over time](https://starchart.cc/incu6us/goimports-reviser.svg)](https://starchart.cc/incu6us/goimports-reviser)

