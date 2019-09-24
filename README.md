# isclib
[![Latest Tag](https://img.shields.io/github/tag/ontariosystems/isclib.svg)](https://github.com/ontariosystems/isclib/tags)
[![CLA assistant](https://cla-assistant.io/readme/badge/ontariosystems/isclib)](https://cla-assistant.io/ontariosystems/isclib)
[![Build Status](https://travis-ci.org/ontariosystems/isclib.svg?branch=master)](https://travis-ci.org/ontariosystems/isclib)
[![Go Report Card](https://goreportcard.com/badge/github.com/ontariosystems/isclib)](https://goreportcard.com/report/github.com/ontariosystems/isclib)
[![GoDoc](https://godoc.org/github.com/ontariosystems/isclib?status.svg)](https://godoc.org/github.com/ontariosystems/isclib)

Go library for interacting with InterSystems Corporation products like Caché, Ensemble, and IRIS Data Platform

It provides methods for determining if ISC products are installed and for interacting with instances of them

### Checking for available ISC commands

```go
package main

import (
	"github.com/ontariosystems/isclib"
)

func main() {
    if isclib.AvailableCommands().Has(isclib.CControlCommand) {
        // perform actions if Cache/Ensemble is installed
    }
    
    if isclib.AvailableCommands().Has(isclib.IrisCommand) {
        //perform actions if Iris is installed
    }
}
```

### Interacting with an instance

You can get access to an instance, find information about the instance (installation directory, status, ports, version, etc.) and perform operations like starting/stopping the instance and executing code in a namespace

```go
package main

import (
	"bytes"
	"fmt"

	"github.com/ontariosystems/isclib"
)

const (
	c = `MAIN
 write $zversion
 do $system.Process.Terminate($job,0)
 quit

`
)

func main() {
	i, err := isclib.LoadInstance("docker")
	if err != nil {
		panic(err)
	}

	if i.Status == "down" {
		if err := i.Start(); err != nil {
			panic(err)
		}
	}

	r := bytes.NewReader([]byte(c))
	if out, err := i.Execute("%SYS", r); err != nil {
		panic(err)
	} else {
		fmt.Println(out)
	}
}
``` 