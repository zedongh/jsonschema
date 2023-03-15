module github.com/santhosh-tekuri/jsonschema/cmd/jv

go 1.15

require (
	github.com/santhosh-tekuri/jsonschema/v5 v5.1.1
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/santhosh-tekuri/jsonschema/v5 => github.com/zedongh/jsonschema/v5 v5.0.0-20230315070017-bc60f9b1b2b7
