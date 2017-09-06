# crypto-dash

> Cryptocurrency information and charts displayed in a dashboard from your terminal.

<img src="./screenshot.png" width="750">

# Usage

build:

```bash
cd example/
go build main.go
```

run:

```bash
./main {coin}

# example
./main ethereum
```

running from within a program:

```go
package main

import (
  "github.com/miguelmota/crypto-dash"
)

func main() {
	cryptodash.Render("ethereum")
}
```

# License

MIT


