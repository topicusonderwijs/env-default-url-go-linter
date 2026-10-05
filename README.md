# Env default URL Go linter

Prevent default URLs meant for development from leaking into code.

A Go analyzer that reports `env-default` struct tags whose value names a single deployment: a hostname or URL, a mail address, a bucket with a deployment marker, or a credential. Such a default is inherited by every deployment that does not set the value itself, so it silently points at whichever environment the code happened to name.

## Usage

### GitHub Actions

```yaml
- uses: actions/checkout@v6
- uses: topicusonderwijs/env-default-url-go-linter@v0.0.1
```

Inputs: `version` (default `v0.0.1`, the latest release; a tag, branch or commit), `packages` (default `./...`), `args` (extra flags) and `working-directory`. The step fails when a finding is reported.

### Locally

```sh
go install github.com/topicusonderwijs/env-default-url-go-linter/cmd/configlint@v0.0.1
configlint ./...
# or as a vet tool
go vet -vettool=$(which configlint) ./...
```

Flags: `-tag` (struct tag holding the default, `env-default` by default), `-directive` (name in the exemption comment) and `-tests` (also check `_test.go` files).

## Exempting a field

A value that is genuinely identical in every deployment can be exempted. The reason is mandatory, and an exemption on a field that is no longer flagged is reported as stale.

```go
//configlint:allow central IdP, identical in every deployment
TokenURL string `env:"TOKEN_URL" env-default:"https://..."`
```

## Development

```sh
go test ./...
```
