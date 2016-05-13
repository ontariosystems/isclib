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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/ontariosystems/isclib"
)

var _ = Describe("Instance", func() {
	Context("InstanceFromQList", func() {
		Context("Invalid qlist", func() {
			instance, err := isclib.InstanceFromQList("1^2^3^4^5^6^7^8")
			It("Returns an error", func() {
				Expect(err).To(HaveOccurred())
			})
			It("Does not return an instance", func() {
				Expect(instance).To(BeNil())
			})
		})

		Context("Valid qlist", func() {
			instance, err := isclib.InstanceFromQList("INSTTEST^/ensemble/instances/insttest/^2015.2.2.805.0.16216^running, since Fri May 13 22:07:02 2016^cache.cpf^56772^57772^62972^ok^")
			It("Does not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("Populates the instance properly from a qlist", func() {
				Expect(instance).ToNot(BeNil())
				Expect(instance.Name).To(Equal("INSTTEST"), "name")
				Expect(instance.Directory).To(Equal("/ensemble/instances/insttest/"), "directory")
				Expect(instance.Version).To(Equal("2015.2.2.805.0.16216"), "version")
				Expect(instance.Status).To(Equal(isclib.InstanceStatusRunning), "status")
				Expect(instance.Activity).To(Equal("since Fri May 13 22:07:02 2016"), "activity")
				Expect(instance.CPFFileName).To(Equal("cache.cpf"), "cpf")
				Expect(instance.SuperServerPort).To(Equal(56772), "ss port")
				Expect(instance.WebServerPort).To(Equal(57772), "ws port")
				Expect(instance.JDBCPort).To(Equal(62972), "jdbc port")
				Expect(instance.State).To(Equal("ok"), "state")
			})
		})
	})
})
