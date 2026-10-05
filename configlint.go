// Package configlint reports configuration defaults that differ per deployment.
//
// A default value in a configuration struct is inherited by every deployment that does not set the
// value itself. For a host, an endpoint, a bucket or a credential that is wrong: the deployment then
// silently runs against whichever environment this repository happened to name. The analyzer reports
// such defaults, and accepts an exemption for a value that is genuinely identical everywhere:
//
//	//configlint:allow central IdP, identical in every deployment
//	TokenURL string `env:"..." env-default:"https://..."`
//
// The exemption must state a reason, so a reviewer can judge it in the diff next to the value it
// excuses. An exemption on a field that is no longer flagged is reported too, which keeps exemptions
// from piling up unnoticed.
package configlint

import (
	"go/ast"
	"net/netip"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = `report configuration defaults that differ per deployment

Flags the value of a struct tag (env-default by default) when it names one deployment: a hostname, a
mail address, a bucket carrying a deployment marker, or a field holding a credential. Exempt a field
that is identical in every deployment with //configlint:allow <reason>.`

// Analyzer reports environment-specific defaults in struct tags.
var Analyzer = &analysis.Analyzer{
	Name:     "configlint",
	Doc:      doc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

var (
	defaultTag   = "env-default"
	directive    = "configlint"
	includeTests bool
)

func init() {
	Analyzer.Flags.StringVar(&defaultTag, "tag", defaultTag,
		"struct tag holding the default value, for projects not using cleanenv")
	Analyzer.Flags.StringVar(&directive, "directive", directive,
		"name used in the //<directive>:allow <reason> exemption comment")
	Analyzer.Flags.BoolVar(&includeTests, "tests", includeTests,
		"also report defaults in _test.go files, where fixtures legitimately carry them")
}

// Hosts that resolve to the same machine in every deployment and may therefore keep a default.
var localHosts = map[string]bool{
	"localhost": true,
	"127.0.0.1": true,
	"0.0.0.0":   true,
	"::1":       true,
}

// Suffixes that make a bare value a hostname rather than, say, a NATS subject or a bucket list.
var hostSuffixes = map[string]bool{
	"net": true, "nl": true, "com": true, "org": true,
	"io": true, "dev": true, "build": true, "education": true,
	"eu": true, "de": true, "be": true, "uk": true, "fr": true,
	"internal": true, "local": true, "cloud": true, "k8s": true,
}

// Tokens that name a single deployment, so a value containing one cannot be right everywhere.
var deploymentMarkers = map[string]bool{
	"test": true, "acc": true, "acceptatie": true, "sandbox": true,
	"staging": true, "prod": true, "productie": true, "dev": true,
}

// Field name suffixes that hold a secret or a client identity, which must never come from a default.
var credentialSuffixes = []string{"clientid", "secret", "key", "password", "token"}

func run(pass *analysis.Pass) (any, error) {
	inspected := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Generated code cannot be fixed at the source, and an exemption in it is wiped on regeneration.
	generated := make(map[string]bool)
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			generated[pass.Fset.Position(file.Pos()).Filename] = true
		}
	}

	inspected.WithStack([]ast.Node{(*ast.StructType)(nil)}, func(node ast.Node, push bool, stack []ast.Node) bool {
		if !push {
			return false
		}
		filename := pass.Fset.Position(node.Pos()).Filename
		if generated[filename] || (!includeTests && strings.HasSuffix(filename, "_test.go")) {
			return false
		}
		for _, field := range node.(*ast.StructType).Fields.List {
			checkField(pass, field, enclosingType(stack))
		}
		return true
	})

	return nil, nil
}

func checkField(pass *analysis.Pass, field *ast.Field, prefix string) {
	// A field without a tag has no default, but may still carry an exemption that is now stale.
	var value string
	var hasDefault bool
	if field.Tag != nil {
		tag, err := strconv.Unquote(field.Tag.Value)
		if err != nil {
			return
		}

		value, hasDefault = reflect.StructTag(tag).Lookup(defaultTag)
	}
	reason, exempted := exemption(field)
	names := fieldNames(field)

	// One tag covers every name in `A, B string`, but the credential rule looks at the name, so each
	// name is judged on its own.
	problems := make(map[string]string, len(names))
	for _, name := range names {
		if problem := problemFor(name, value, hasDefault); problem != "" {
			problems[name] = problem
		}
	}

	declared := prefix + strings.Join(names, ", ")

	switch {
	case len(problems) == 0 && exempted:
		pass.Reportf(field.Pos(), "%s: the //%s:allow exemption is no longer needed, %s",
			declared, directive, describeDefault(value, hasDefault))
	case len(problems) > 0 && !exempted:
		for _, name := range names {
			if problem, found := problems[name]; found {
				pass.Reportf(field.Pos(), "%s: %q is %s; remove the default and set the value per deployment, or exempt the field with //%s:allow <reason>",
					prefix+name, value, problem, directive)
			}
		}
	case len(problems) > 0 && reason == "":
		pass.Reportf(field.Pos(), "%s: the //%s:allow exemption gives no reason; state why this value is identical in every deployment",
			declared, directive)
	}
}

// exemption reports whether the field carries an allow directive, and the reason given for it. Both
// the doc comment above the field and the line comment after it are accepted.
func exemption(field *ast.Field) (string, bool) {
	for _, group := range []*ast.CommentGroup{field.Doc, field.Comment} {
		if group == nil {
			continue
		}
		for _, comment := range group.List {
			text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
			rest, found := strings.CutPrefix(text, directive+":allow")
			if !found || !isDirectiveEnd(rest) {
				continue
			}
			return strings.TrimSpace(strings.TrimLeft(rest, ":-")), true
		}
	}

	return "", false
}

// isDirectiveEnd keeps //configlint:allowance from being read as //configlint:allow plus a reason.
func isDirectiveEnd(rest string) bool {
	return rest == "" || strings.ContainsRune(" \t:-", rune(rest[0]))
}

func problemFor(name, value string, hasDefault bool) string {
	if !hasDefault || value == "" {
		return ""
	}
	if isCredentialField(name) {
		return "a credential with a default"
	}
	if isEnvironmentSpecific(value) {
		return "an environment-specific default"
	}

	return ""
}

func isCredentialField(name string) bool {
	lowered := strings.ToLower(name)
	for _, suffix := range credentialSuffixes {
		if strings.HasSuffix(lowered, suffix) {
			return true
		}
	}

	return false
}

func isEnvironmentSpecific(value string) bool {
	if parsed, err := url.Parse(value); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		return !localHosts[parsed.Hostname()]
	}
	if addr, ok := ipOf(value); ok {
		return !addr.IsLoopback() && !addr.IsUnspecified()
	}
	if address, domain, found := strings.Cut(value, "@"); found && address != "" && isHostname(domain) {
		return true
	}
	if isHostname(value) {
		return !localHosts[hostOf(value)]
	}

	return hasDeploymentMarker(value)
}

// isHostname keeps NATS subjects and bucket lists, which also contain dots, from being read as hosts.
func isHostname(value string) bool {
	host := hostOf(value)
	if localHosts[host] {
		return true
	}

	labels := strings.Split(host, ".")
	if len(labels) < 2 {
		return false
	}

	return hostSuffixes[strings.ToLower(labels[len(labels)-1])]
}

// ipOf reads a bare IP address, with or without a port, such as 10.0.3.4:8080 or [::1]:80.
func ipOf(value string) (netip.Addr, bool) {
	value = withoutPath(value)
	if addr, err := netip.ParseAddr(value); err == nil {
		return addr, true
	}
	if addrPort, err := netip.ParseAddrPort(value); err == nil {
		return addrPort.Addr(), true
	}

	return netip.Addr{}, false
}

// hostOf drops the path, query and port, so that api.example.com:8080/v1 is judged by api.example.com.
func hostOf(value string) string {
	host, _, _ := strings.Cut(withoutPath(value), ":")
	return host
}

func withoutPath(value string) string {
	if end := strings.IndexAny(value, "/?#"); end >= 0 {
		return value[:end]
	}

	return value
}

// hasDeploymentMarker looks for a whole token naming one deployment, so that "development" as the
// deployment default is left alone while a bucket named "my-app-test" is not.
func hasDeploymentMarker(value string) bool {
	isSeparator := func(r rune) bool {
		return strings.ContainsRune("._-/: ", r)
	}
	for _, token := range strings.FieldsFunc(strings.ToLower(value), isSeparator) {
		if deploymentMarkers[token] {
			return true
		}
	}

	return false
}

// enclosingType names the struct type a field belongs to, so that a diagnostic reads FQDN.Form rather
// than just Form. An anonymous struct has no name to report, and then the field name stands alone; it
// must not borrow the name of the type it is nested in, or it would read like a field of that type.
func enclosingType(stack []ast.Node) string {
	if len(stack) >= 2 {
		if spec, ok := stack[len(stack)-2].(*ast.TypeSpec); ok {
			return spec.Name.Name + "."
		}
	}

	return ""
}

func fieldNames(field *ast.Field) []string {
	if len(field.Names) == 0 {
		return []string{"embedded field"}
	}

	names := make([]string, len(field.Names))
	for i, name := range field.Names {
		names[i] = name.Name
	}

	return names
}

func describeDefault(value string, hasDefault bool) string {
	if !hasDefault {
		return "the field has no default"
	}

	return strconv.Quote(value) + " is not environment-specific"
}
