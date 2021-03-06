/*
 * Licensed to Elasticsearch under one or more contributor
 * license agreements. See the NOTICE file distributed with
 * this work for additional information regarding copyright
 * ownership. Elasticsearch licenses this file to you under
 * the Apache License, Version 2.0 (the "License"); you may
 * not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/elastic/go-elasticsearch/generator/api"
)

// Method is the test generator for a given method. It contains multiple test specs, each of which contains one or
// more tests.
type Method struct {
	Specs    []*testSpec
	Method   *api.Method
	template *template.Template
}

// NewMethod instantiates a tester for a given API namespace.
func NewMethod(specDir, methodName string, methods map[string]*api.Method, templates *template.Template) (*Method, error) {
	m := &Method{
		Specs: []*testSpec{},
	}
	var ok bool
	if m.Method, ok = methods[methodName]; !ok {
		var name string
		switch methodName {
		case "ingest":
			name = "ingest.put_pipeline"
		case "mlt":
			fallthrough
		case "search.aggregation":
			fallthrough
		case "search.highlight":
			fallthrough
		case "search.inner_hits":
			name = "search"
		default:
			return nil, fmt.Errorf("invalid method name: %s", methodName)
		}
		if m.Method, ok = methods[name]; !ok {
			return nil, fmt.Errorf("invalid method name: %s", name)
		}
	}
	testSpecDir := filepath.Join(specDir, "test", methodName)
	files, err := ioutil.ReadDir(testSpecDir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".yaml" {
			continue
		}
		ts, err := newTestSpec(filepath.Join(testSpecDir, file.Name()), methods, m.Method.Name, templates)
		if err != nil {
			return nil, err
		}
		m.Specs = append(m.Specs, ts)
	}
	if m.template = templates.Lookup("test.tmpl"); m.template == nil {
		return nil, fmt.Errorf("cannot find template for tests")
	}
	return m, nil
}

// Generate generates the test file for a method.
func (m *Method) Generate(outputDir string) error {
	goFile, err := m.Method.NewWriter(outputDir, m.Method.TestFileName)
	if err != nil {
		return fmt.Errorf("failed to create test writer for %s: %s", m.Method.Name, err)
	}
	if err = m.template.Execute(goFile, m); err != nil {
		return err
	}
	return nil
}
