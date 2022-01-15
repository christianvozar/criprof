# criprof
Container Runtime Interface profiling and introspection. Useful for tracking down containers in logs or grouping by runtime characteristics.

criprof looks for hints about the running container and its runtime. The aim is to provide an inventory of the executing container and its runtime for debugging purposes. 

## Usage

```Go
package main

import (
	"fmt"

	"github.com/christianvozar/criprof"
)

func main() {
	i := criprof.New()

	fmt.Println(i.JSON())
}
```

## Contribution

If you are aware of additional hints or profile information worth surfacing please open an issue and I'll add it to the package.
