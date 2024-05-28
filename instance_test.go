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
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/user"
	"syscall"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Instance", func() {
	const (
		cacheqlist    = "INSTTEST^/ensemble/instances/insttest/^2015.2.2.805.0.16216^running, since Fri May 13 22:07:02 2016^cache.cpf^56772^57772^62972^ok^"
		durableqlist  = "INSTTEST^/ensemble/instances/insttest/^2015.2.2.805.0.16216^running, since Fri May 13 22:07:02 2016^cache.cpf^56772^57772^62972^ok^^^^/mgr/config"
		ensembleqlist = "INSTTEST^/ensemble/instances/insttest/^2015.2.2.805.0.16216^running, since Fri May 13 22:07:02 2016^cache.cpf^56772^57772^62972^ok^Ensemble"
		irisqlist     = "INSTTEST^/ensemble/instances/insttest/^2018.1.1.643.0^running, since Fri May 13 22:07:02 2016^iris.cpf^56772^57772^62972^ok^IRIS^^^/mgr/config"
		legacyqlist   = "INSTTEST^/ensemble/instances/insttest/^2012.2.3.903.2.12515^down, last used Thu Sep 15 18:58:30 2016^cache.cpf^56772^57772^62972"
		mirrorqlist   = "INSTTEST^/ensemble/instances/insttest/^2015.2.2.805.0.16216^running, since Fri May 13 22:07:02 2016^cache.cpf^56772^57772^62972^ok^^Failover^Primary^/mgr/config"
		warnqlist     = "INSTTEST^/ensemble/instances/insttest/^2015.2.2.805.0.16216^running, since Fri May 13 22:07:02 2016^cache.cpf^56772^57772^62972^warn^"
	)
	var (
		instance               *Instance
		err                    error
		timeout                time.Duration
		origCSessionCommand    string
		origIrisSessionCommand string
	)

	BeforeEach(func() {
		// make just enough of a parameters.isc to be able to find the manager user
		parameterReader = func(directory string, file string) (io.ReadCloser, error) {
			u, err := user.Current()
			if err != nil {
				return nil, err
			}
			g, err := user.LookupGroupId(u.Gid)
			if err != nil {
				return nil, err
			}

			parametersContent := fmt.Sprintf("security_settings.manager_user: %s\nsecurity_settings.manager_group: %s", u.Username, g.Name)
			return io.NopCloser(bytes.NewBufferString(parametersContent)), nil
		}
	})

	Describe("InstanceFromQList", func() {
		Context("Invalid qlist", func() {
			BeforeEach(func() {
				instance, err = InstanceFromQList("1^2^3^4^5^6^7")
			})
			It("Returns an error", func() {
				Expect(err).To(HaveOccurred())
			})
			It("Does not return an instance", func() {
				Expect(instance).To(BeNil())
			})
		})

		Context("Valid old qlist", func() {
			BeforeEach(func() {
				instance, err = InstanceFromQList(legacyqlist)
			})
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
			BeforeEach(func() {
				instance, err = InstanceFromQList(warnqlist)
			})
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
			BeforeEach(func() {
				instance, err = InstanceFromQList(durableqlist)
			})
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
		Context("Valid IRIS qlist", func() {
			BeforeEach(func() {
				instance, err = InstanceFromQList(irisqlist)
			})
			It("Does not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("Populates the instance properly from a qlist", func() {
				Expect(instance).ToNot(BeNil())
				Expect(instance.Name).To(Equal("INSTTEST"), "name")
				Expect(instance.Directory).To(Equal("/ensemble/instances/insttest/"), "directory")
				Expect(instance.Version).To(Equal("2018.1.1.643.0"), "version")
				Expect(instance.Status).To(Equal(InstanceStatusRunning), "status")
				Expect(instance.Activity).To(Equal("since Fri May 13 22:07:02 2016"), "activity")
				Expect(instance.CPFFileName).To(Equal("iris.cpf"), "cpf")
				Expect(instance.SuperServerPort).To(Equal(56772), "ss port")
				Expect(instance.WebServerPort).To(Equal(57772), "ws port")
				Expect(instance.JDBCPort).To(Equal(62972), "jdbc port")
				Expect(instance.State).To(Equal("ok"), "state")
				Expect(instance.Product).To(Equal(Iris), "product")
				Expect(instance.DataDirectory).To(Equal("/mgr/config"), "data directory")
			})
		})
		Context("Valid Mirroring", func() {
			BeforeEach(func() {
				instance, err = InstanceFromQList(mirrorqlist)
			})
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

	Describe("DetermineISCDatFileName", func() {
		Context("The product is Cache", func() {
			It("Returns the correct DAT filename", func() {
				instance, _ = InstanceFromQList(cacheqlist)
				Expect(instance.DetermineISCDatFileName()).To(Equal(CacheDatName))
			})
		})
		Context("The product is Ensemble", func() {
			It("Returns the correct DAT filename", func() {
				instance, _ = InstanceFromQList(ensembleqlist)
				Expect(instance.DetermineISCDatFileName()).To(Equal(CacheDatName))
			})
		})
		Context("The product is Iris", func() {
			It("Returns the correct DAT filename", func() {
				instance, _ = InstanceFromQList(irisqlist)
				Expect(instance.DetermineISCDatFileName()).To(Equal(IrisDatName))
			})
		})
	})
	Describe("SessionCommand", func() {
		BeforeEach(func() {
			origCSessionCommand = CSessionPath()
			SetCSessionPath("/somepath/csession")
			origIrisSessionCommand = IrisSessionCommand()
			SetIrisSessionCommand("/somepath/iris session")
		})
		AfterEach(func() {
			SetCSessionPath(origCSessionCommand)
			SetIrisSessionCommand(origIrisSessionCommand)
		})
		Describe("The product is Cache", func() {
			BeforeEach(func() {
				instance, _ = InstanceFromQList(cacheqlist)
			})
			Context("with no namespace or command", func() {
				It("Returns the correct command to execute", func() {
					cmd := instance.SessionCommand("", "")
					Expect(cmd.Path).To(Equal("/somepath/csession"))
					Expect(cmd.Args).To(BeEquivalentTo([]string{"/somepath/csession", "INSTTEST"}))
				})
			})
			Context("with a namespace and command", func() {
				It("Returns the correct command to execute", func() {
					cmd := instance.SessionCommand("TEST", "TEST^TEST")
					Expect(cmd.Path).To(Equal("/somepath/csession"))
					Expect(cmd.Args).To(BeEquivalentTo([]string{"/somepath/csession", "INSTTEST", "-U", "TEST", "TEST^TEST"}))
				})
			})
			Context("with a different session command", func() {
				BeforeEach(func() {
					origCSessionCommand = CSessionPath()
					SetCSessionPath("dsession")
				})
				AfterEach(func() {
					SetCSessionPath(origCSessionCommand)
				})
				It("Returns the correct command to execute", func() {
					cmd := instance.SessionCommand("TEST", "TEST^TEST")
					Expect(cmd.Path).To(Equal("dsession"))
					Expect(cmd.Args).To(BeEquivalentTo([]string{"dsession", "INSTTEST", "-U", "TEST", "TEST^TEST"}))
				})
			})
			Context("with a execution user configured", func() {
				BeforeEach(func() {
					instance.executionSysProcAttr = &syscall.SysProcAttr{
						Credential: &syscall.Credential{
							Uid: uint32(0),
							Gid: uint32(0),
						},
					}
				})
				It("Returns the correct command to execute", func() {
					cmd := instance.SessionCommand("TEST", "TEST^TEST")
					Expect(cmd.Path).To(Equal("/somepath/csession"))
					Expect(cmd.Args).To(BeEquivalentTo([]string{"/somepath/csession", "INSTTEST", "-U", "TEST", "TEST^TEST"}))
					Expect(cmd.SysProcAttr.Credential.Uid).To(Equal(uint32(0)))
					Expect(cmd.SysProcAttr.Credential.Gid).To(Equal(uint32(0)))
				})
			})
		})
		Describe("The product is Ensemble", func() {
			BeforeEach(func() {
				instance, _ = InstanceFromQList(ensembleqlist)
			})
			Context("with no namespace or command", func() {
				It("Returns the correct command to execute", func() {
					cmd := instance.SessionCommand("", "")
					Expect(cmd.Path).To(Equal("/somepath/csession"))
					Expect(cmd.Args).To(BeEquivalentTo([]string{"/somepath/csession", "INSTTEST"}))
				})
			})
			Context("with a namespace and command", func() {
				It("Returns the correct command to execute", func() {
					cmd := instance.SessionCommand("TEST", "TEST^TEST")
					Expect(cmd.Path).To(Equal("/somepath/csession"))
					Expect(cmd.Args).To(BeEquivalentTo([]string{"/somepath/csession", "INSTTEST", "-U", "TEST", "TEST^TEST"}))
				})
			})
			Context("with a different session command", func() {
				BeforeEach(func() {
					origCSessionCommand = CSessionPath()
					SetCSessionPath("dsession")
				})
				AfterEach(func() {
					SetCSessionPath(origCSessionCommand)
				})
				It("Returns the correct command to execute", func() {
					cmd := instance.SessionCommand("TEST", "TEST^TEST")
					Expect(cmd.Path).To(Equal("dsession"))
					Expect(cmd.Args).To(BeEquivalentTo([]string{"dsession", "INSTTEST", "-U", "TEST", "TEST^TEST"}))
				})
			})
			Context("with a execution user configured", func() {
				BeforeEach(func() {
					instance.executionSysProcAttr = &syscall.SysProcAttr{
						Credential: &syscall.Credential{
							Uid: uint32(0),
							Gid: uint32(0),
						},
					}
				})
				It("Returns the correct command to execute", func() {
					cmd := instance.SessionCommand("TEST", "TEST^TEST")
					Expect(cmd.Path).To(Equal("/somepath/csession"))
					Expect(cmd.Args).To(BeEquivalentTo([]string{"/somepath/csession", "INSTTEST", "-U", "TEST", "TEST^TEST"}))
					Expect(cmd.SysProcAttr.Credential.Uid).To(Equal(uint32(0)))
					Expect(cmd.SysProcAttr.Credential.Gid).To(Equal(uint32(0)))
				})
			})
		})
		Describe("The product is Iris", func() {
			BeforeEach(func() {
				instance, _ = InstanceFromQList(irisqlist)
			})
			Context("with no namespace or command", func() {
				It("Returns the correct command to execute", func() {
					cmd := instance.SessionCommand("", "")
					Expect(cmd.Path).To(Equal("/somepath/iris"))
					Expect(cmd.Args).To(BeEquivalentTo([]string{"/somepath/iris", "session", "INSTTEST"}))
				})
			})
			Context("with a namespace and command", func() {
				It("Returns the correct command to execute", func() {
					cmd := instance.SessionCommand("TEST", "TEST^TEST")
					Expect(cmd.Path).To(Equal("/somepath/iris"))
					Expect(cmd.Args).To(BeEquivalentTo([]string{"/somepath/iris", "session", "INSTTEST", "-U", "TEST", "TEST^TEST"}))
				})
			})
			Context("with a namespace and command", func() {
				BeforeEach(func() {
					globalIrisSessionCommand = "dsession"
				})
				It("Returns the correct command to execute", func() {
					cmd := instance.SessionCommand("TEST", "TEST^TEST")
					Expect(cmd.Path).To(Equal("dsession"))
					Expect(cmd.Args).To(BeEquivalentTo([]string{"dsession", "INSTTEST", "-U", "TEST", "TEST^TEST"}))
				})
			})
			Context("with multiple sub-commands", func() {
				BeforeEach(func() {
					globalIrisSessionCommand = "iris session please sir"
				})
				It("Returns the correct command to execute", func() {
					cmd := instance.SessionCommand("TEST", "TEST^TEST")
					Expect(cmd.Path).To(Equal("iris"))
					Expect(cmd.Args).To(BeEquivalentTo([]string{"iris", "session", "please", "sir", "INSTTEST", "-U", "TEST", "TEST^TEST"}))
				})
			})
			Context("with a execution user configured", func() {
				BeforeEach(func() {
					instance.executionSysProcAttr = &syscall.SysProcAttr{
						Credential: &syscall.Credential{
							Uid: uint32(0),
							Gid: uint32(0),
						},
					}
				})
				It("Returns the correct command to execute", func() {
					cmd := instance.SessionCommand("TEST", "TEST^TEST")
					Expect(cmd.Path).To(Equal("/somepath/iris"))
					Expect(cmd.Args).To(BeEquivalentTo([]string{"/somepath/iris", "session", "INSTTEST", "-U", "TEST", "TEST^TEST"}))
					Expect(cmd.SysProcAttr.Credential.Uid).To(Equal(uint32(0)))
					Expect(cmd.SysProcAttr.Credential.Gid).To(Equal(uint32(0)))
				})
			})
		})
	})
	Describe("LicenseKeyFilePath", func() {
		Context("The product is Cache", func() {
			It("Returns the correct DAT filename", func() {
				instance, _ = InstanceFromQList(cacheqlist)
				Expect(instance.LicenseKeyFilePath()).To(Equal("/ensemble/instances/insttest/mgr/cache.key"))
			})
		})
		Context("The product is Ensemble", func() {
			It("Returns the correct DAT filename", func() {
				instance, _ = InstanceFromQList(ensembleqlist)
				Expect(instance.LicenseKeyFilePath()).To(Equal("/ensemble/instances/insttest/mgr/cache.key"))
			})
		})
		Context("The product is Iris", func() {
			It("Returns the correct DAT filename", func() {
				instance, _ = InstanceFromQList(irisqlist)
				Expect(instance.LicenseKeyFilePath()).To(Equal("/mgr/config/mgr/license.key"))
			})
		})
	})
	Describe("WaitForReady", func() {
		Context("With timeout", func() {
			Context("Does not come up", func() {
				BeforeEach(func() {
					instance, _ = InstanceFromQList(legacyqlist)
					ctx, can := context.WithTimeout(context.Background(), 50*time.Millisecond)
					defer can()
					err = instance.WaitForReady(ctx)
				})
				It("Returns an error", func() {
					Expect(err).To(HaveOccurred())
				})
				It("Timed out", func() {
					Expect(err).Should(MatchError(context.DeadlineExceeded))
				})
			})
			Context("Does come up", func() {
				BeforeEach(func() {
					timeout = 500 * time.Millisecond
					getQlist = func(instanceName string, _ *syscall.SysProcAttr) (string, error) {
						return legacyqlist, nil
					}
					time.AfterFunc(timeout/2, func() {
						getQlist = func(instanceName string, _ *syscall.SysProcAttr) (string, error) {
							return durableqlist, nil
						}
					})
					instance, _ = InstanceFromQList(legacyqlist)
					ctx, can := context.WithTimeout(context.Background(), timeout)
					defer can()
					err = instance.WaitForReady(ctx)
				})
				It("Does not return an error", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})
	Describe("sessionCommand", func() {
		Describe("The product is Cache", func() {
			BeforeEach(func() {
				instance, _ = InstanceFromQList(cacheqlist)
			})
			Context("with overridden session command", func() {
				BeforeEach(func() {
					instance.SessionPath = "dsession"
				})
				It("returns the overridden session path", func() {
					Expect(instance.sessionCommand()).To(Equal(instance.SessionPath))
				})
			})
			Context("with default session command", func() {
				It("returns the default session command", func() {
					Expect(instance.sessionCommand()).To(Equal(globalCSessionPath))
				})
			})
		})
		Describe("The product is Ensemble", func() {
			BeforeEach(func() {
				instance, _ = InstanceFromQList(ensembleqlist)
			})
			Context("with overridden session command", func() {
				BeforeEach(func() {
					instance.SessionPath = "dsession"
				})
				It("returns the overridden session path", func() {
					Expect(instance.sessionCommand()).To(Equal(instance.SessionPath))
				})
			})
			Context("with default session command", func() {
				It("returns the default session command", func() {
					Expect(instance.sessionCommand()).To(Equal(globalCSessionPath))
				})
			})
		})
		Describe("The product is Iris", func() {
			BeforeEach(func() {
				instance, _ = InstanceFromQList(irisqlist)
			})
			Context("with overridden session command", func() {
				BeforeEach(func() {
					instance.SessionPath = "dsession"
				})
				It("returns the overridden session path", func() {
					Expect(instance.sessionCommand()).To(Equal(instance.SessionPath))
				})
			})
			Context("with default session command", func() {
				It("returns the default session command", func() {
					Expect(instance.sessionCommand()).To(Equal(globalIrisSessionCommand))
				})
			})
		})
	})
	Describe("controlPath", func() {
		Describe("The product is Cache", func() {
			BeforeEach(func() {
				instance, _ = InstanceFromQList(cacheqlist)
			})
			Context("with overridden contorl command", func() {
				BeforeEach(func() {
					instance.ControlPath = "dcontrol"
				})
				It("returns the overridden control path", func() {
					Expect(instance.controlPath()).To(Equal(instance.ControlPath))
				})
			})
			Context("with default control command", func() {
				It("returns the default control command", func() {
					Expect(instance.controlPath()).To(Equal(globalCControlPath))
				})
			})
		})
		Describe("The product is Ensemble", func() {
			BeforeEach(func() {
				instance, _ = InstanceFromQList(ensembleqlist)
			})
			Context("with overridden contorl command", func() {
				BeforeEach(func() {
					instance.ControlPath = "dcontrol"
				})
				It("returns the overridden control path", func() {
					Expect(instance.controlPath()).To(Equal(instance.ControlPath))
				})
			})
			Context("with default control command", func() {
				It("returns the default control command", func() {
					Expect(instance.controlPath()).To(Equal(globalCControlPath))
				})
			})
		})
		Describe("The product is Iris", func() {
			BeforeEach(func() {
				instance, _ = InstanceFromQList(irisqlist)
			})
			Context("with overridden control command", func() {
				BeforeEach(func() {
					instance.ControlPath = "jris"
				})
				It("returns the overridden session path", func() {
					Expect(instance.controlPath()).To(Equal(instance.ControlPath))
				})
			})
			Context("with default session command", func() {
				It("returns the default session command", func() {
					Expect(instance.controlPath()).To(Equal(globalIrisPath))
				})
			})
		})
	})

	Describe("Update", func() {
		// To test instance updates when running somewhere that doesn't actually have access to the
		// parameters.isc file, such as `iscenv` wrapping `csession` or `iris`
		Context("Valid qlist without parameters.isc", func() {
			BeforeEach(func() {
				parameterReader = func(directory string, file string) (io.ReadCloser, error) {
					return nil, os.ErrNotExist
				}
				instance, err = InstanceFromQList(cacheqlist)
				Expect(err).ToNot(HaveOccurred())
			})

			It("Does not return an error", func() {
				err := instance.Update()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
