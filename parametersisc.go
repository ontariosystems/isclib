/*
Copyright 2016 Ontario Systems

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

package isclib

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const (
	parameterLinePattern = `^\s*(?:([^.]+)\.)?(.+):\s*(.*)$`
)

var (
	parameterLineRegexp = regexp.MustCompile(parameterLinePattern)
)

// ParametersISC represents the contents of the parameters ISC file
type ParametersISC map[string]ParametersISCGroup

// ParametersISCGroup represents a group of related parameters from the ISC file
type ParametersISCGroup map[string]*ParametersISCEntry

// ParametersISCEntry represents a single entry from the parameters ISC file
type ParametersISCEntry struct {
	// The group for this entry (the portion of the key before the .)
	Group string

	// The name of this entry (the portion of the key after the .)
	Name string

	// The values for this entry
	Values []string
}

// LoadParametersISC will load the parameters contained in the provided reader
// It returns the ParametersISC data structure and any error encountered
func LoadParametersISC(r io.Reader) (ParametersISC, error) {
	pi := make(ParametersISC)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		t := scanner.Text()
		if strings.TrimSpace(t) == "" {
			continue
		}
		m := parameterLineRegexp.FindStringSubmatch(t)
		if len(m) != 4 {
			return nil, fmt.Errorf("malformed parameter line: %s", t)
		}

		group := m[1]
		name := m[2]
		value := m[3]
		if _, ok := pi[group]; !ok {
			pi[group] = make(ParametersISCGroup)
		}

		if _, ok := pi[group][name]; !ok {
			pi[group][name] = &ParametersISCEntry{
				Group:  group,
				Name:   name,
				Values: make([]string, 0),
			}
		}

		pi[group][name].Values = append(pi[group][name].Values, value)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return pi, nil
}

// Values will, given a set of identifiers making up a parameter key, return the values at that key
// Identifiers can be...
//
//   - the full key (group.name)
//   - the group, name as two separate parameters
//   - A single parameter representing the name of a parameter in the "" group
//
// It returns the values at that key or an empty slice if it does not exist
func (pi ParametersISC) Values(identifiers ...string) []string {
	var group, name string
	switch len(identifiers) {
	case 1:
		s := strings.Split(identifiers[0], ".")
		if len(s) == 1 {
			group = ""
			name = s[0]
		} else {
			group = s[0]
			name = s[1]
		}
	case 2:
		group = identifiers[0]
		name = identifiers[1]
	default:
		return []string{}
	}

	g := pi[group]
	if g == nil {
		return []string{}
	}

	e := pi[group][name]
	if e == nil {
		return []string{}
	}

	v := pi[group][name].Values
	if v == nil {
		return []string{}
	}

	return v
}

// Value will, given a set of identifiers making up a parameter key, return the single value at that key
// Identifiers can be...
//
// - the full key (group.name)
// - the group, name as two separate parameters
// - A single parameter representing the name of a parameter in the "" group
//
// It returns the value if a single value exists for the key or "" if it does not
func (pi ParametersISC) Value(identifiers ...string) string {
	values := pi.Values(identifiers...)
	if len(values) == 1 {
		return values[0]
	}

	return ""
}

// Key returns the full group.name key for this element
func (pie ParametersISCEntry) Key() string {
	if pie.Group == "" {
		return pie.Name
	}

	return pie.Group + "." + pie.Name
}
