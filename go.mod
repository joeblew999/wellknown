module github.com/joeblew999/wellknown

go 1.25.3

// Development tools (automatically installed with `go get`)
tool (
	github.com/air-verse/air // Hot-reload dev server
	github.com/snonky/pocketbase-gogen // Type-safe PB code generator
)

require (
	github.com/santhosh-tekuri/jsonschema/v5 v5.3.1 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
	golang.org/x/text v0.14.0 // indirect
)
