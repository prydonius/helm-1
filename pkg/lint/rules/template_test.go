/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rules

import (
	"k8s.io/helm/pkg/lint/support"
	"strings"
	"testing"
)

const templateTestBasedir = "./testdata/albatross"

func TestValidateAllowedExtension(t *testing.T) {
	var failTest = []string{"/foo", "/test.yml", "/test.toml", "test.yml"}
	for _, test := range failTest {
		err := validateAllowedExtension(test)
		if err == nil || !strings.Contains(err.Error(), "needs to use .yaml or .tpl extension") {
			t.Errorf("validateAllowedExtension('%s') to return \"needs to use .yaml or .tpl extension\", got no error", test)
		}
	}
	var successTest = []string{"/foo.yaml", "foo.yaml", "foo.tpl", "/foo/bar/baz.yaml"}
	for _, test := range successTest {
		err := validateAllowedExtension(test)
		if err != nil {
			t.Errorf("validateAllowedExtension('%s') to return no error but got \"%s\"", test, err.Error())
		}
	}
}

func TestValidateQuotes(t *testing.T) {
	// add `| quote` lint error
	var failTest = []string{"foo: {{.Release.Service }}", "foo:  {{.Release.Service }}", "- {{.Release.Service }}", "foo: {{default 'Never' .restart_policy}}", "-  {{.Release.Service }} "}

	for _, test := range failTest {
		err := validateQuotes("testTemplate.yaml", test)
		if err == nil || !strings.Contains(err.Error(), "use the sprig \"quote\" function") {
			t.Errorf("validateQuotes('%s') to return \"use the sprig \"quote\" function:\", got no error.", test)
		}
	}

	var successTest = []string{"foo: {{.Release.Service | quote }}", "foo:  {{.Release.Service | quote }}", "- {{.Release.Service | quote }}", "foo: {{default 'Never' .restart_policy | quote }}", "foo: \"{{ .Release.Service }}\"", "foo: \"{{ .Release.Service }} {{ .Foo.Bar }}\"", "foo: \"{{ default 'Never' .Release.Service }} {{ .Foo.Bar }}\"", "foo:  {{.Release.Service | squote }}"}

	for _, test := range successTest {
		err := validateQuotes("testTemplate.yaml", test)
		if err != nil {
			t.Errorf("validateQuotes('%s') to return not error and got \"%s\"", test, err.Error())
		}
	}

	// Surrounding quotes
	failTest = []string{"foo: {{.Release.Service }}-{{ .Release.Bar }}", "foo: {{.Release.Service }} {{ .Release.Bar }}", "- {{.Release.Service }}-{{ .Release.Bar }}", "- {{.Release.Service }}-{{ .Release.Bar }} {{ .Release.Baz }}", "foo: {{.Release.Service | default }}-{{ .Release.Bar }}"}

	for _, test := range failTest {
		err := validateQuotes("testTemplate.yaml", test)
		if err == nil || !strings.Contains(err.Error(), "Wrap your substitution functions in quotes") {
			t.Errorf("validateQuotes('%s') to return \"Wrap your substitution functions in quotes\", got no error", test)
		}
	}

}

func TestTemplate(t *testing.T) {
	linter := support.Linter{ChartDir: templateTestBasedir}
	Templates(&linter)
	res := linter.Messages

	if len(res) != 1 {
		t.Fatalf("Expected one error, got %d, %v", len(res), res)
	}

	if !strings.Contains(res[0].Text, "deliberateSyntaxError") {
		t.Errorf("Unexpected error: %s", res[0])
	}
}
