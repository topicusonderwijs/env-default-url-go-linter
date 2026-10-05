package testfile

// Fixture carries a default that is reported only when test files are included.
type Fixture struct {
	Endpoint string `env-default:"https://api.test.example.com"`
}
