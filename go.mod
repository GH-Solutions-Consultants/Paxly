module github.com/GH-Solutions-Consultants/Paxly

go 1.21.4

require (
	github.com/GH-Solutions-Consultants/Paxly/core v0.0.0

	// Require plugin modules with pseudo-versions
	github.com/GH-Solutions-Consultants/Paxly/plugins/go_plugin v0.0.0-00010101000000-000000000000
	github.com/GH-Solutions-Consultants/Paxly/plugins/javascript_plugin v0.0.0-00010101000000-000000000000
	github.com/GH-Solutions-Consultants/Paxly/plugins/python_plugin v0.0.0-00010101000000-000000000000
	github.com/GH-Solutions-Consultants/Paxly/plugins/rust_plugin v0.0.0-00010101000000-000000000000
	github.com/go-playground/validator/v10 v10.22.1
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.7.0
	github.com/stretchr/testify v1.8.4
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/Masterminds/semver/v3 v3.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/crypto v0.19.0 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace github.com/GH-Solutions-Consultants/Paxly/core => ./core

replace github.com/GH-Solutions-Consultants/Paxly/plugins/go_plugin => ./plugins/go

replace github.com/GH-Solutions-Consultants/Paxly/plugins/javascript_plugin => ./plugins/javascript

replace github.com/GH-Solutions-Consultants/Paxly/plugins/python_plugin => ./plugins/python

replace github.com/GH-Solutions-Consultants/Paxly/plugins/rust_plugin => ./plugins/rust
