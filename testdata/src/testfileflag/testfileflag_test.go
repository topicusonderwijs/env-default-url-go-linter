package testfileflag

// Fixture carries a default that is reported only when test files are included.
type Fixture struct {
	Endpoint string `env-default:"https://api.test.example.com"` // want `Fixture.Endpoint: "https://api.test.example.com" is an environment-specific default`
}
