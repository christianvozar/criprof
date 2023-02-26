# criprof

Container Runtime Interface profiling and introspection. CRI profiling and introspection can be used to gain insight into the behavior and performance of containerized applications, and diagnose and troubleshoot issues related to the containers. Typically, while runtime environments will introspect their own profile information if a stack utilizes multiple runtime environemnts it can be useful to detect which environment is being utilized at runtime. This information can be used to optimize the performance of the containerized applications, improve the reliability and stability of the containerized infrastructure, and to identify potential security vulnerabilities.

criprof uses hints much like the ohai project about the running container and its runtime. The aim is to provide an inventory of the executing container and its runtime for debugging purposes. 

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
