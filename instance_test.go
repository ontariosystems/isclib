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

package isclib

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("Instance", func() {
	Context("InstanceFromQList", func() {
		Context("Invalid qlist", func() {
			instance, err := InstanceFromQList("1^2^3^4^5^6^7")
			It("Returns an error", func() {
				Expect(err).To(HaveOccurred())
			})
			It("Does not return an instance", func() {
				Expect(instance).To(BeNil())
			})
		})

		Context("Valid old qlist", func() {
			instance, err := InstanceFromQList("INSTTEST^/ensemble/instances/insttest/^2012.2.3.903.2.12515^down, last used Thu Sep 15 18:58:30 2016^cache.cpf^56772^57772^62972")
			It("Does not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("Populates the instance properly from a qlist", func() {
				Expect(instance).ToNot(BeNil())
				Expect(instance.Name).To(Equal("INSTTEST"), "name")
				Expect(instance.Directory).To(Equal("/ensemble/instances/insttest/"), "directory")
				Expect(instance.Version).To(Equal("2012.2.3.903.2.12515"), "version")
				Expect(instance.Status).To(Equal(InstanceStatusDown), "status")
				Expect(instance.Activity).To(Equal("last used Thu Sep 15 18:58:30 2016"), "activity")
				Expect(instance.CPFFileName).To(Equal("cache.cpf"), "cpf")
				Expect(instance.SuperServerPort).To(Equal(56772), "ss port")
				Expect(instance.WebServerPort).To(Equal(57772), "ws port")
				Expect(instance.JDBCPort).To(Equal(62972), "jdbc port")
				Expect(instance.State).To(Equal("ok"), "state")
			})
		})
		Context("Valid qlist", func() {
			instance, err := InstanceFromQList("INSTTEST^/ensemble/instances/insttest/^2015.2.2.805.0.16216^running, since Fri May 13 22:07:02 2016^cache.cpf^56772^57772^62972^warn^")
			It("Does not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("Populates the instance properly from a qlist", func() {
				Expect(instance).ToNot(BeNil())
				Expect(instance.Name).To(Equal("INSTTEST"), "name")
				Expect(instance.Directory).To(Equal("/ensemble/instances/insttest/"), "directory")
				Expect(instance.Version).To(Equal("2015.2.2.805.0.16216"), "version")
				Expect(instance.Status).To(Equal(InstanceStatusRunning), "status")
				Expect(instance.Activity).To(Equal("since Fri May 13 22:07:02 2016"), "activity")
				Expect(instance.CPFFileName).To(Equal("cache.cpf"), "cpf")
				Expect(instance.SuperServerPort).To(Equal(56772), "ss port")
				Expect(instance.WebServerPort).To(Equal(57772), "ws port")
				Expect(instance.JDBCPort).To(Equal(62972), "jdbc port")
				Expect(instance.State).To(Equal("warn"), "state")
				Expect(instance.DataDirectory).To(Equal("/ensemble/instances/insttest/"), "data directory")
			})
		})
		Context("Valid Durable %SYS qlist", func() {
			instance, err := InstanceFromQList("INSTTEST^/ensemble/instances/insttest/^2015.2.2.805.0.16216^running, since Fri May 13 22:07:02 2016^cache.cpf^56772^57772^62972^ok^^^^/mgr/config")
			It("Does not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("Populates the instance properly from a qlist", func() {
				Expect(instance).ToNot(BeNil())
				Expect(instance.Name).To(Equal("INSTTEST"), "name")
				Expect(instance.Directory).To(Equal("/ensemble/instances/insttest/"), "directory")
				Expect(instance.Version).To(Equal("2015.2.2.805.0.16216"), "version")
				Expect(instance.Status).To(Equal(InstanceStatusRunning), "status")
				Expect(instance.Activity).To(Equal("since Fri May 13 22:07:02 2016"), "activity")
				Expect(instance.CPFFileName).To(Equal("cache.cpf"), "cpf")
				Expect(instance.SuperServerPort).To(Equal(56772), "ss port")
				Expect(instance.WebServerPort).To(Equal(57772), "ws port")
				Expect(instance.JDBCPort).To(Equal(62972), "jdbc port")
				Expect(instance.State).To(Equal("ok"), "state")
				Expect(instance.Product).To(Equal(Cache), "product")
				Expect(instance.DataDirectory).To(Equal("/mgr/config"), "data directory")
			})
		})
		Context("Valid Mirroring", func() {
			instance, err := InstanceFromQList("INSTTEST^/ensemble/instances/insttest/^2015.2.2.805.0.16216^running, since Fri May 13 22:07:02 2016^cache.cpf^56772^57772^62972^ok^^Failover^Primary^/mgr/config")
			It("Does not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("Populates the instance properly from a qlist", func() {
				Expect(instance).ToNot(BeNil())
				Expect(instance.Name).To(Equal("INSTTEST"), "name")
				Expect(instance.Directory).To(Equal("/ensemble/instances/insttest/"), "directory")
				Expect(instance.Version).To(Equal("2015.2.2.805.0.16216"), "version")
				Expect(instance.Status).To(Equal(InstanceStatusRunning), "status")
				Expect(instance.Activity).To(Equal("since Fri May 13 22:07:02 2016"), "activity")
				Expect(instance.CPFFileName).To(Equal("cache.cpf"), "cpf")
				Expect(instance.SuperServerPort).To(Equal(56772), "ss port")
				Expect(instance.WebServerPort).To(Equal(57772), "ws port")
				Expect(instance.JDBCPort).To(Equal(62972), "jdbc port")
				Expect(instance.State).To(Equal("ok"), "state")
				Expect(instance.Product).To(Equal(Cache), "product")
				Expect(instance.MirrorMemberType).To(Equal("Failover"), "mirror type")
				Expect(instance.MirrorStatus).To(Equal("Primary"), "mirror status")
				Expect(instance.DataDirectory).To(Equal("/mgr/config"), "data directory")
			})
		})
	})
	Context("WaitForReady", func() {
		var i int
		BeforeEach(func() {
			i = 0
		})
		Context("With timeout", func() {
			Context("Does not come up", func() {
				instance, err := InstanceFromQList("INSTTEST^/ensemble/instances/insttest/^2012.2.3.903.2.12515^down, last used Thu Sep 15 18:58:30 2016^cache.cpf^56772^57772^62972")
				ctx, can := context.WithTimeout(context.Background(), 50*time.Millisecond)
				defer can()
				err = instance.WaitForReady(ctx)
				It("Returns an error", func() {
					Expect(err).To(HaveOccurred())
				})
				It("Timed out", func() {
					Expect(err).Should(MatchError(context.DeadlineExceeded))
				})
			})
			Context("Does come up", func() {
				getQlist = func(instanceName string) (string, error) {
					if i >= 3 {
						return "INSTTEST^/ensemble/instances/insttest/^2015.2.2.805.0.16216^running, since Fri May 13 22:07:02 2016^cache.cpf^56772^57772^62972^ok^^^^/mgr/config", nil
					}
					i++
					return "INSTTEST^/ensemble/instances/insttest/^2012.2.3.903.2.12515^down, last used Thu Sep 15 18:58:30 2016^cache.cpf^56772^57772^62972", nil
				}
				instance, err := InstanceFromQList("INSTTEST^/ensemble/instances/insttest/^2012.2.3.903.2.12515^down, last used Thu Sep 15 18:58:30 2016^cache.cpf^56772^57772^62972")
				ctx, can := context.WithTimeout(context.Background(), 500*time.Millisecond)
				defer can()
				err = instance.WaitForReady(ctx)
				It("Does not return an error", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})
})
