package probe

// Base is embedded in Probe.
type Base struct{}

// Probe pins down both the reported and the deliberately unreported shapes: the values after Bucket
// contain dots or deployment-like words but name no deployment, and must stay silent.
type Probe struct {
	//configlint:allow central IdP, identical in every deployment
	Allowed string `env-default:"https://login.test.example.com/idp/oauth2/token"`

	//configlint:allow
	AllowedWithoutReason string `env-default:"https://login.test.example.com"` // want `Probe.AllowedWithoutReason: the //configlint:allow exemption gives no reason`

	//configlint:allow no longer needed
	StaleExemption string `env-default:"localhost"` // want `Probe.StaleExemption: the //configlint:allow exemption is no longer needed, "localhost" is not environment-specific`

	//configlint:allowance stuff
	NotADirective string `env-default:"https://api.test.example.com"` // want `Probe.NotADirective: "https://api.test.example.com" is an environment-specific default`

	//configlint:allow: colon form, identical in every deployment
	ColonForm string `env-default:"https://api.test.example.com"`

	Endpoint    string `env-default:"https://api.test.example.com"` // want `Probe.Endpoint: "https://api.test.example.com" is an environment-specific default`
	BareHost    string `env-default:"storage.example.nl:8082"`      // want `Probe.BareHost: "storage.example.nl:8082" is an environment-specific default`
	MailAddress string `env-default:"no-reply@test.example.com"`    // want `Probe.MailAddress: "no-reply@test.example.com" is an environment-specific default`
	Bucket      string `env-default:"my-app-test"`                  // want `Probe.Bucket: "my-app-test" is an environment-specific default`
	ClientID    string `env-default:"Probe-ably-not-a-secret"`      // want `Probe.ClientID: "Probe-ably-not-a-secret" is a credential with a default`

	HostWithPath string `env-default:"api.example.com/v1"`           // want `Probe.HostWithPath: "api.example.com/v1" is an environment-specific default`
	IPAddress    string `env-default:"10.0.3.4:8080"`                // want `Probe.IPAddress: "10.0.3.4:8080" is an environment-specific default`
	InternalHost string `env-default:"db.internal:5432"`             // want `Probe.InternalHost: "db.internal:5432" is an environment-specific default`
	GermanHost   string `env-default:"service.example.de"`           // want `Probe.GermanHost: "service.example.de" is an environment-specific default`
	ClusterHost  string `env-default:"db.default.svc.cluster.local"` // want `Probe.ClusterHost: "db.default.svc.cluster.local" is an environment-specific default`

	LoopbackIP    string `env-default:"127.0.0.1:8080"`
	LoopbackIPv6  string `env-default:"::1"`
	UnspecifiedIP string `env-default:"0.0.0.0:80"`

	Base `env-default:"https://api.test.example.com"` // want `Probe.embedded field: "https://api.test.example.com" is an environment-specific default`

	// Every name is judged on its own: only APIKey is a credential.
	Host, APIKey string `env-default:"cache"` // want `Probe.APIKey: "cache" is a credential with a default`

	//configlint:allow shared by all, identical in every deployment
	ExemptA, ExemptB string `env-default:"https://api.test.example.com"`

	//configlint:allow no longer needed
	StaleA, StaleB string `env-default:"cache"` // want `Probe.StaleA, StaleB: the //configlint:allow exemption is no longer needed, "cache" is not environment-specific`

	//configlint:allow tag was removed
	NoTagExempted string // want `Probe.NoTagExempted: the //configlint:allow exemption is no longer needed, the field has no default`

	NoTag string

	LocalURL       string `env-default:"http://localhost:8081/"`
	Deployment     string `env-default:"development"`
	Subject        string `env-default:"orders.mail.send"`
	MetricsBuckets string `env-default:".001, .01, .05, .10"`
	Region         string `env-default:"eu-west-3"`
	BucketName     string `env-default:"bucket_name"`
	ClientName     string `env-default:"my-app-http-client"`
	EmptyKey       string `env-default:""`
	MigrationPath  string `env-default:"/bin/migrations"`
	NoDefault      string `env:"SOME_URL"`
	Port           int    `env-default:"4444"`
}

// Unreachable is not referenced from any root config struct, which is exactly what the reflect-based
// variant cannot see and the analyzer can.
type Unreachable struct {
	Endpoint string `env-default:"https://api.sandbox.example.com"` // want `Unreachable.Endpoint: "https://api.sandbox.example.com" is an environment-specific default`
}

// Nested has an anonymous struct, whose fields are reported without borrowing the name Nested.
type Nested struct {
	Host  string `env-default:"localhost"`
	Inner struct {
		Host string `env-default:"https://api.test.example.com"` // want `^Host: "https://api.test.example.com" is an environment-specific default`
	}
}
