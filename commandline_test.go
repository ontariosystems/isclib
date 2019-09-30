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
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ontariosystems/isclib"
)

var _ = Describe("Commands", func() {
	var (
		origCControlCommand string
		origCSessionCommand string
		origIrisCommand     string
		origPath            string
		tempCControlCommand *os.File
		tempCSessionCommand *os.File
		tempIrisCommand     *os.File
	)

	Context("AvailableCommands", func() {
		BeforeEach(func() {
			origPath = os.Getenv("PATH")
			os.Setenv("PATH", origPath+":/tmp")
			origCControlCommand = isclib.CControlPath()
			isclib.SetCControlPath("/somepath/ccontrol")
			origCSessionCommand = isclib.CSessionPath()
			isclib.SetCSessionPath("/somepath/csession")
			origIrisCommand = isclib.IrisPath()
			isclib.SetIrisPath("/somepath/iris")
		})
		AfterEach(func() {
			os.Setenv("PATH", origPath)
			isclib.SetCControlPath(origCControlCommand)
			isclib.SetCSessionPath(origCSessionCommand)
			isclib.SetIrisPath(origIrisCommand)
		})
		Context("No ISC executables", func() {
			It("should return NoCommand", func() {
				commands := isclib.AvailableCommands()
				Expect(commands).To(Equal(isclib.NoCommand), "no command")
				Expect(commands.Has(isclib.CControlCommand)).To(BeFalse())
				Expect(commands.Has(isclib.CSessionCommand)).To(BeFalse())
				Expect(commands.Has(isclib.IrisCommand)).To(BeFalse())
			})
		})
		Context("With ISC executables", func() {
			BeforeEach(func() {
				var err error
				if tempCControlCommand, err = ioutil.TempFile(os.TempDir(), "ccontrol"); err != nil {
					panic(err)
				}
				os.Chmod(tempCControlCommand.Name(), 0755)
				if tempCSessionCommand, err = ioutil.TempFile(os.TempDir(), "csession"); err != nil {
					panic(err)
				}
				os.Chmod(tempCSessionCommand.Name(), 0755)
				if tempIrisCommand, err = ioutil.TempFile(os.TempDir(), "iris"); err != nil {
					panic(err)
				}
				os.Chmod(tempIrisCommand.Name(), 0755)
			})
			AfterEach(func() {
				os.Remove(tempCControlCommand.Name())
				os.Remove(tempCSessionCommand.Name())
				os.Remove(tempIrisCommand.Name())
			})
			Context("only iris", func() {
				BeforeEach(func() {
					origIrisCommand = isclib.IrisPath()
					isclib.SetIrisPath(tempIrisCommand.Name())
				})
				AfterEach(func() {
					isclib.SetIrisPath(origIrisCommand)
				})
				It("should return only iris", func() {
					commands := isclib.AvailableCommands()
					Expect(commands).To(Equal(isclib.IrisCommand), "iris only")
					Expect(commands.Has(isclib.CControlCommand)).To(BeFalse())
					Expect(commands.Has(isclib.CSessionCommand)).To(BeFalse())
					Expect(commands.Has(isclib.IrisCommand)).To(BeTrue())
					Expect(commands.Has(isclib.NoCommand)).To(BeFalse())
				})
			})
			Context("only cache commands", func() {
				BeforeEach(func() {
					origCControlCommand = isclib.CControlPath()
					isclib.SetCControlPath(tempCControlCommand.Name())
					origCSessionCommand = isclib.CSessionPath()
					isclib.SetCSessionPath(tempCSessionCommand.Name())
				})
				AfterEach(func() {
					isclib.SetCControlPath(origCControlCommand)
					isclib.SetCSessionPath(origCSessionCommand)
				})
				It("should return ccontrol and csession", func() {
					commands := isclib.AvailableCommands()
					Expect(commands).To(Equal(isclib.CControlCommand|isclib.CSessionCommand), "ccontrol and csession")
					Expect(commands.Has(isclib.CControlCommand)).To(BeTrue())
					Expect(commands.Has(isclib.CSessionCommand)).To(BeTrue())
					Expect(commands.Has(isclib.IrisCommand)).To(BeFalse())
					Expect(commands.Has(isclib.NoCommand)).To(BeFalse())
				})
			})
			Context("both iris and cache commands", func() {
				BeforeEach(func() {
					origCControlCommand = isclib.CControlPath()
					isclib.SetCControlPath(tempCControlCommand.Name())
					origCSessionCommand = isclib.CSessionPath()
					isclib.SetCSessionPath(tempCSessionCommand.Name())
					origIrisCommand = isclib.IrisPath()
					isclib.SetIrisPath(tempIrisCommand.Name())
				})
				AfterEach(func() {
					isclib.SetCControlPath(origCControlCommand)
					isclib.SetCSessionPath(origCSessionCommand)
					isclib.SetIrisPath(origIrisCommand)
				})
				It("should return all ISC commands", func() {
					commands := isclib.AvailableCommands()
					Expect(commands).To(Equal(isclib.CControlCommand|isclib.CSessionCommand|isclib.IrisCommand), "all commands")
					Expect(commands.Has(isclib.CControlCommand)).To(BeTrue())
					Expect(commands.Has(isclib.CSessionCommand)).To(BeTrue())
					Expect(commands.Has(isclib.IrisCommand)).To(BeTrue())
					Expect(commands.Has(isclib.NoCommand)).To(BeFalse())
				})
			})
		})
	})

	Context("Set", func() {
		var commands isclib.Commands
		BeforeEach(func() {
			commands = isclib.NoCommand
		})

		Context("Adding a command", func() {
			It("includes the added command", func() {
				commands.Set(isclib.IrisCommand)
				Expect(commands.Has(isclib.IrisCommand)).To(BeTrue())
			})
		})
		Context("Adding a command that already exists", func() {
			BeforeEach(func() {
				commands = isclib.IrisCommand
			})
			It("still includes the added command", func() {
				commands.Set(isclib.IrisCommand)
				Expect(commands.Has(isclib.IrisCommand)).To(BeTrue())
			})
		})
	})

	Context("Clear", func() {
		var commands isclib.Commands
		BeforeEach(func() {
			commands = isclib.IrisCommand
		})

		Context("Clearing a command", func() {
			It("does not include the cleared command", func() {
				commands.Clear(isclib.IrisCommand)
				Expect(commands.Has(isclib.IrisCommand)).To(BeFalse())
			})
		})
	})

	Context("Has", func() {
		var commands isclib.Commands
		Context("Has set commands", func() {
			BeforeEach(func() {
				commands = isclib.CControlCommand | isclib.CSessionCommand
			})
			It("has the set commands", func() {
				Expect(commands.Has(isclib.CControlCommand)).To(BeTrue())
				Expect(commands.Has(isclib.CSessionCommand)).To(BeTrue())
			})
			It("does not have the unset commands", func() {
				Expect(commands.Has(isclib.IrisCommand)).To(BeFalse())
			})
		})
	})
})
