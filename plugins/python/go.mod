// plugins/python/go.mod
module github.com/GH-Solutions-Consultants/Paxly/plugins/python

go 1.21.4

require (
	github.com/GH-Solutions-Consultants/Paxly/core v0.0.0
	github.com/sirupsen/logrus v1.9.3
)

replace github.com/GH-Solutions-Consultants/Paxly/core => ../../core
// Removed the following line as it's incorrect:
// replace github.com/GH-Solutions-Consultants/Paxly/plugins/python_plugin => ./