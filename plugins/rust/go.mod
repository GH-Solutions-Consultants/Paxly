module github.com/GH-Solutions-Consultants/Paxly/plugins/rust

go 1.21.4

require (
	github.com/GH-Solutions-Consultants/Paxly/core v0.0.0
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/Masterminds/semver/v3 v3.3.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.22.1 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/crypto v0.19.0 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/GH-Solutions-Consultants/Paxly/core => ../../core

replace github.com/GH-Solutions-Consultants/Paxly/plugins/rust => ./
