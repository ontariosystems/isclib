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

package isclib_test

import (
	"bytes"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/ontariosystems/isclib/v2"
)

type failReader struct{}

func (*failReader) Read(p []byte) (n int, err error) { return 0, fmt.Errorf("Blam!") }

var _ = Describe("ParametersISC", func() {
	Context("LoadParametersISC", func() {
		Context("Failing reader", func() {
			r := new(failReader)
			_, err := isclib.LoadParametersISC(r)
			It("Returns an error", func() {
				Expect(err).To(MatchError("Blam!"))
			})
		})
		Context("Invalid data", func() {
			r := bytes.NewBufferString(`

ng: ngval
g1.n1: g1n1val
nope

`)
			_, err := isclib.LoadParametersISC(r)
			It("Returns an error", func() {
				Expect(err).To(MatchError("malformed parameter line: nope"))
			})
		})
		Context("Valid data", func() {
			r := bytes.NewBufferString(`

ng: ngval
g1.n1: g1n1val
g1.n2: g1n2val
g2.n1: g2n1val
g2.n2: 
g2.n3:
dup: dup1
dup: dup2
dup: dup3

`)
			pi, err := isclib.LoadParametersISC(r)
			It("Does not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("Contains the appropriate values", func() {
				Expect(pi[""]["ng"]).To(Equal(&isclib.ParametersISCEntry{Group: "", Name: "ng", Values: []string{"ngval"}}))
				Expect(pi["g1"]["n1"]).To(Equal(&isclib.ParametersISCEntry{Group: "g1", Name: "n1", Values: []string{"g1n1val"}}))
				Expect(pi["g1"]["n2"]).To(Equal(&isclib.ParametersISCEntry{Group: "g1", Name: "n2", Values: []string{"g1n2val"}}))
				Expect(pi["g2"]["n1"]).To(Equal(&isclib.ParametersISCEntry{Group: "g2", Name: "n1", Values: []string{"g2n1val"}}))
				Expect(pi["g2"]["n2"]).To(Equal(&isclib.ParametersISCEntry{Group: "g2", Name: "n2", Values: []string{""}}))
				Expect(pi["g2"]["n3"]).To(Equal(&isclib.ParametersISCEntry{Group: "g2", Name: "n3", Values: []string{""}}))
				Expect(pi[""]["dup"]).To(Equal(&isclib.ParametersISCEntry{Group: "", Name: "dup", Values: []string{"dup1", "dup2", "dup3"}}))
			})
		})
	})

	Context("Value", func() {
		pi := make(isclib.ParametersISC)
		pi[""] = make(isclib.ParametersISCGroup)
		pi[""]["ng"] = &isclib.ParametersISCEntry{Group: "", Name: "ng", Values: []string{"ngval"}}
		pi[""]["dup"] = &isclib.ParametersISCEntry{Group: "", Name: "ng", Values: []string{"dup1", "dup2"}}
		pi["g1"] = make(isclib.ParametersISCGroup)
		pi["g1"]["n1"] = &isclib.ParametersISCEntry{Group: "g1", Name: "n1", Values: []string{"g1n1val"}}
		pi["g1"]["n2"] = &isclib.ParametersISCEntry{Group: "g1", Name: "n1", Values: nil} // This one is malformed but there's always the change that someone but it incorrectly themselves
		pi["g3"] = nil                                                                    // This one is malformed but there's always the change that someone but it incorrectly themselves

		It("Looks up single values as expected", func() {
			Expect(pi.Value("ng")).To(Equal("ngval"))
			Expect(pi.Value("dup")).To(Equal(""))
			Expect(pi.Value("g1.n1")).To(Equal("g1n1val"))
			Expect(pi.Value("g1", "n1")).To(Equal("g1n1val"))
			Expect(pi.Value("g1", "n2")).To(Equal(""))
			Expect(pi.Value("g1", "n3")).To(Equal(""))
			Expect(pi.Value("g2", "n1")).To(Equal(""))
			Expect(pi.Value("g2.n1")).To(Equal(""))
			Expect(pi.Value("g1", "n1", "x", "y")).To(Equal(""))
		})
	})

	Context("Values", func() {
		pi := make(isclib.ParametersISC)
		pi[""] = make(isclib.ParametersISCGroup)
		pi[""]["ng"] = &isclib.ParametersISCEntry{Group: "", Name: "ng", Values: []string{"ngval"}}
		pi[""]["dup"] = &isclib.ParametersISCEntry{Group: "", Name: "ng", Values: []string{"dup1", "dup2"}}
		pi["g1"] = make(isclib.ParametersISCGroup)
		pi["g1"]["n1"] = &isclib.ParametersISCEntry{Group: "g1", Name: "n1", Values: []string{"g1n1val"}}
		pi["g1"]["n2"] = &isclib.ParametersISCEntry{Group: "g1", Name: "n1", Values: nil} // This one is malformed but there's always the change that someone but it incorrectly themselves

		It("Looks up values as expected", func() {
			Expect(pi.Values("ng")).To(Equal([]string{"ngval"}))
			Expect(pi.Values("dup")).To(Equal([]string{"dup1", "dup2"}))
			Expect(pi.Values("g1.n1")).To(Equal([]string{"g1n1val"}))
			Expect(pi.Values("g1", "n1")).To(Equal([]string{"g1n1val"}))
			Expect(pi.Values("g1", "n2")).To(Equal([]string{}))
			Expect(pi.Values("g1", "n3")).To(Equal([]string{}))
			Expect(pi.Values("g2", "n1")).To(Equal([]string{}))
			Expect(pi.Values("g2.n1")).To(Equal([]string{}))
			Expect(pi.Values("g1", "n1", "x", "y")).To(Equal([]string{}))
		})
	})

	Context("Key", func() {
		It("Returns the appropriate key for an entry", func() {
			Expect(isclib.ParametersISCEntry{Group: "", Name: "ng", Values: []string{"ngval"}}.Key()).To(Equal("ng"))
			Expect(isclib.ParametersISCEntry{Group: "g1", Name: "n1", Values: []string{"g1n1val"}}.Key()).To(Equal("g1.n1"))
		})
	})
})
