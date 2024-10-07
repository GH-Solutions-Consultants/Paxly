// main.go
package main

import (
	"github.com/GH-Solutions-Consultants/Paxly/cmd"
	"github.com/GH-Solutions-Consultants/Paxly/core"
	_ "github.com/GH-Solutions-Consultants/Paxly/plugins/go_plugin"
	_ "github.com/GH-Solutions-Consultants/Paxly/plugins/javascript_plugin"
	_ "github.com/GH-Solutions-Consultants/Paxly/plugins/python"        // Updated import path
	_ "github.com/GH-Solutions-Consultants/Paxly/plugins/rust_plugin"
)

func main() {
	core.InitializeLogger() // Initialize the logger first
	cmd.Execute()
}
