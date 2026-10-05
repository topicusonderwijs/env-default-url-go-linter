// Command configlint reports configuration defaults that differ per deployment.
//
// Run it standalone or as a vet tool:
//
//	configlint ./...
//	go vet -vettool=$(which configlint) ./...
package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/topicusonderwijs/env-default-url-go-linter"
)

func main() {
	singlechecker.Main(configlint.Analyzer)
}
