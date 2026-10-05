package customtag

// Only the tag named by the -tag flag holds the default.
type Config struct {
	Endpoint string `default:"https://api.test.example.com"` // want `Config.Endpoint: "https://api.test.example.com" is an environment-specific default`
	Ignored  string `env-default:"https://api.test.example.com"`
}
