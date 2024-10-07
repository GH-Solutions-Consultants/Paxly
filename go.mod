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

    // Updated plugin dependencies with new versions
    github.com/GH-Solutions-Consultants/Paxly/plugins/go_plugin v0.1.3
    github.com/GH-Solutions-Consultants/Paxly/plugins/javascript_plugin v0.1.3
    github.com/GH-Solutions-Consultants/Paxly/plugins/python_plugin v0.1.3
    github.com/GH-Solutions-Consultants/Paxly/plugins/rust_plugin v0.1.3
)
