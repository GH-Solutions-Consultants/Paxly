module github.com/GH-Solutions-Consultants/Paxly

go 1.21

require (
    github.com/Masterminds/semver/v3 v3.2.1
    github.com/go-playground/validator/v10 v10.15.5
    github.com/pkg/errors v0.9.1
    github.com/sirupsen/logrus v1.9.3
    github.com/spf13/cobra v1.7.0
    github.com/stretchr/testify v1.8.4
    gopkg.in/yaml.v2 v2.4.0
    gopkg.in/yaml.v3 v3.0.1
    github.com/GH-Solutions-Consultants/Paxly/core v0.0.0
)

replace github.com/GH-Solutions-Consultants/Paxly/core => ./core
replace github.com/GH-Solutions-Consultants/Paxly/plugins/go_plugin => ./plugins/go
replace github.com/GH-Solutions-Consultants/Paxly/plugins/javascript_plugin => ./plugins/javascript
replace github.com/GH-Solutions-Consultants/Paxly/plugins/python_plugin => ./plugins/python
replace github.com/GH-Solutions-Consultants/Paxly/plugins/rust_plugin => ./plugins/rust
