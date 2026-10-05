package customdirective

// Only the directive named by the -directive flag exempts a field.
type Config struct {
	//lint:allow central IdP, identical in every deployment
	Exempted string `env-default:"https://api.test.example.com"`

	//configlint:allow not the directive in use
	NotExempted string `env-default:"https://api.test.example.com"` // want `Config.NotExempted: "https://api.test.example.com" is an environment-specific default`
}
