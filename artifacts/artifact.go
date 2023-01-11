package artifact

import _ "embed"

//go:embed example.nomad
var ExampleJobspec string

//go:embed example-short.nomad
var ExampleJobspecShort string

//go:embed connect.nomad
var ExampleJobspecConnect string

//go:embed connect-short.nomad
var ExampleJobspecConnectShort string
