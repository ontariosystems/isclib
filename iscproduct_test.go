/*
Copyright 2017 Ontario Systems

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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/ontariosystems/isclib"
)

var _ = Describe("InstanceStatus", func() {
	Context("ParseISCProduct", func() {
		It("Returns a default product of Cache", func() {
			Expect(isclib.ParseISCProduct("")).To(Equal(isclib.Cache), "default product")
			Expect(isclib.ParseISCProduct("NotAProduct")).To(Equal(isclib.Cache), "default product")
		})
		It("Successfully parses Cache as a product", func() {
			Expect(isclib.ParseISCProduct("Cache")).To(Equal(isclib.Cache), "Cache product")
		})
		It("Successfully parses ISC product as a product", func() {
			Expect(isclib.ParseISCProduct("Ensemble")).To(Equal(isclib.Ensemble), "Ensemble product")
		})
		It("Successfully parses IRIS as a product", func() {
			Expect(isclib.ParseISCProduct("IDP")).To(Equal(isclib.Iris), "IRIS product")
		})
	})
})
