package configlint_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/topicusonderwijs/env-default-url-go-linter"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), configlint.Analyzer, "probe")
}

func TestTestFilesAreSkippedByDefault(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), configlint.Analyzer, "testfile")
}

func TestTestFilesFlag(t *testing.T) {
	setFlag(t, "tests", "true")

	analysistest.Run(t, analysistest.TestData(), configlint.Analyzer, "testfileflag")
}

func TestTagFlag(t *testing.T) {
	setFlag(t, "tag", "default")

	analysistest.Run(t, analysistest.TestData(), configlint.Analyzer, "customtag")
}

func TestDirectiveFlag(t *testing.T) {
	setFlag(t, "directive", "lint")

	analysistest.Run(t, analysistest.TestData(), configlint.Analyzer, "customdirective")
}

// setFlag sets an analyzer flag for one test. The flags are package-level state, so the previous value
// is restored afterwards.
func setFlag(t *testing.T, name, value string) {
	t.Helper()

	flag := configlint.Analyzer.Flags.Lookup(name)
	previous := flag.Value.String()
	if err := configlint.Analyzer.Flags.Set(name, value); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := configlint.Analyzer.Flags.Set(name, previous); err != nil {
			t.Fatal(err)
		}
	})
}
