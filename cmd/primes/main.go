package main

import (
	"github.com/zalgonoise/x/cli"
)

var modes = []string{"serve", "build"}

func main() {
	runner := cli.NewRunner("primes",
		cli.WithOneOf(modes...),
		cli.WithExecutors(map[string]cli.Executor{
			"serve": cli.Executable(ExecServe),
			"build": cli.Executable(ExecBuild),
		}),
	)

	cli.Run(runner)
}
