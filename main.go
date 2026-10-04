package main

import (
	"os"

	"k8s.io/component-base/cli"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"

	"github.com/negin/hello-politecnico-scheduler/pkg/plugins/hello"
)

func main() {
	command := app.NewSchedulerCommand(
		app.WithPlugin(hello.Name, hello.New),
	)
	code := cli.Run(command)
	os.Exit(code)
}
