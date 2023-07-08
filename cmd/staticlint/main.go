/*
Staticlint uses to union some great analyzers.

List of analyzers used in this package:
  - printf.Analyzer, check consistency of Printf format strings and arguments.
  - shadow.Analyzer, check for possible unintended shadowing of variables.
  - structtag.Analyzer, checks struct field tags are well-formed.
  - bodyclose.Analyzer, checks that all opened bodies are closed.
  - exit.Analyzer, checks using os.Exit() in main function of main package

To run Staticlint you need to build this file and run it, like you running usual binary file.
In args you need to provide file that you want to check.
*/
package main

import (
	"github.com/VladimirMovsesyan/praktikum-devops/internal/analyzers/exit"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
	"strings"
)

func main() {
	// Variable checks contains analyzers
	checks := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		bodyclose.Analyzer,
		exit.Analyzer,
	}

	for _, v := range staticcheck.Analyzers {
		if strings.Contains(v.Analyzer.Name, "SA") {
			checks = append(checks, v.Analyzer)
		}
	}

	multichecker.Main(
		checks...,
	)
}
