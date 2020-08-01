# go-majordomo

[![Tag](https://img.shields.io/github/tag/wealdtech/go-majordomo.svg)](https://github.com/wealdtech/go-majordomo/releases/)
[![License](https://img.shields.io/github/license/wealdtech/go-majordomo.svg)](LICENSE)
[![GoDoc](https://godoc.org/github.com/wealdtech/go-majordomo?status.svg)](https://godoc.org/github.com/wealdtech/go-majordomo)
[![Travis CI](https://img.shields.io/travis/wealdtech/go-majordomo.svg)](https://travis-ci.org/wealdtech/go-majordomo)
[![codecov.io](https://img.shields.io/codecov/c/github/wealdtech/go-majordomo.svg)](https://codecov.io/github/wealdtech/go-majordomo)
[![Go Report Card](https://goreportcard.com/badge/github.com/wealdtech/go-majordomo)](https://goreportcard.com/report/github.com/wealdtech/go-majordomo)

Central access to resources, locally or from secret managers.


## Table of Contents

- [Install](#install)
- [Usage](#usage)
- [Maintainers](#maintainers)
- [Contribute](#contribute)
- [License](#license)

## Install

`go-majordomo` is a standard Go module which can be installed with:

```sh
go get github.com/wealdtech/go-majordomo
```

## Usage

Majordomo manages _confidants_.  A confidant is a module that holds secrets that can be accessed through a custom URL.  Confidants includes in this module are:
  - `direct` secrets that are simple values
  - `file` secrets that are held in a named file
  - `asm` secrets that are stored on Amazon secrets manager
  - `gsm` secrets that are stored on Google secrets manager

Details about how to configure each confidant are in the relevant confidant's go docs.

Creating new confidants should be a relatively simple task; all that is required is to implement the `Confidant` interface.

Majordomo itself is defined as an interface.  This is to allow more complicated implementations (load balancing, retries, caching _etc._) if required.  The standard implementation is in 'standard'

### Example

#### Fetching a secret using the file confidant.
```go
package main

import (
	"context"
	"fmt"

	"github.com/wealdtech/go-majordomo/confidants/file"
	standardmajordomo "github.com/wealdtech/go-majordomo/standard"
)

func main() {
	ctx := context.Background()
	// Create the majordomo service.
	service, err := standardmajordomo.New(ctx)
	if err != nil {
		panic(err)
	}

	// Create and register the file confidant.
	confidant, err := file.New(ctx)
	if err != nil {
		panic(err)
	}
	err = service.RegisterConfidant(ctx, confidant)
	if err != nil {
		panic(err)
	}

	// Fetch a value from the service.
	value, err := service.Fetch(ctx, "file:///home/me/secrets/password.txt")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Value is %s\n", string(value))
}
```

## Maintainers

Jim McDonald: [@mcdee](https://github.com/mcdee).

## Contribute

Contributions welcome. Please check out [the issues](https://github.com/wealdtech/go-majordomo/issues).

## License

[Apache-2.0](LICENSE) © 2019 Weald Technology Trading Ltd
