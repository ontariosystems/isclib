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
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/ontariosystems/isclib/v2"
)

var _ = Describe("ImportDescription", func() {
	Context("NewImportDescription", func() {
		It("Only allows one **", func() {
			_, err := isclib.NewImportDescription("/a/b/c/**/d/e/f/**/*.xml", "")
			Expect(err).To(MatchError(isclib.ErrTooManyRecursiveDirs))
		})

		It("Does not allow * in the directory", func() {
			var err error
			_, err = isclib.NewImportDescription("/a/b/c/*/d/e/f/**/*.xml", "")
			Expect(err).To(MatchError(isclib.ErrWildcardInDirectory))

			_, err = isclib.NewImportDescription("/a/b/c/*/d/e/f/*.xml", "")
			Expect(err).To(MatchError(isclib.ErrWildcardInDirectory))
		})

		It("Does not allow trailing paths between a ** and the file pattern", func() {
			_, err := isclib.NewImportDescription("/a/b/c/**/d/e/f/*.xml", "")
			Expect(err).To(MatchError(isclib.ErrPathAfterRecursiveDirs))
		})

		It("Ensures that there is a path separator between ** and the trailing file pattern", func() {
			_, err := isclib.NewImportDescription("/a/b/c/***.xml", "")
			Expect(err).To(MatchError(isclib.ErrMissingPathSeparator))
		})

		It("Properly parses the directory and file pattern", func() {
			var id *isclib.ImportDescription
			var err error
			var cwd string

			if cwd, err = os.Getwd(); err != nil {
				panic(err)
			}

			id, err = isclib.NewImportDescription("*.xml", "")
			Expect(err).To(Not(HaveOccurred()))
			Expect(id.Dir).To(Equal(cwd))
			Expect(id.FilePattern).To(Equal("*.xml"))
			Expect(id.Recursive).To(BeFalse())

			id, err = isclib.NewImportDescription("/a/b/c/*.xml", "")
			Expect(err).To(Not(HaveOccurred()))
			Expect(id.Dir).To(Equal("/a/b/c"))
			Expect(id.FilePattern).To(Equal("*.xml"))
			Expect(id.Recursive).To(BeFalse())

			id, err = isclib.NewImportDescription("**/*.xml", "")
			Expect(err).To(Not(HaveOccurred()))
			Expect(id.Dir).To(Equal(cwd))
			Expect(id.FilePattern).To(Equal("*.xml"))
			Expect(id.Recursive).To(BeTrue())

			id, err = isclib.NewImportDescription("/a/b/c/**/*.xml", "")
			Expect(err).To(Not(HaveOccurred()))
			Expect(id.Dir).To(Equal("/a/b/c"))
			Expect(id.FilePattern).To(Equal("*.xml"))
			Expect(id.Recursive).To(BeTrue())
		})

		It("Creates the correct import string", func() {
			var id *isclib.ImportDescription
			var err error

			id, err = isclib.NewImportDescription("/a/b/c/*.xml", "/t1")
			Expect(err).To(Not(HaveOccurred()))
			Expect(id.String()).To(Equal(`##class(%SYSTEM.OBJ).ImportDir("/a/b/c","*.xml","/t1",,0)`))

			id, err = isclib.NewImportDescription("/a/b/c/**/abc.xml", "/t2")
			Expect(err).To(Not(HaveOccurred()))
			Expect(id.String()).To(Equal(`##class(%SYSTEM.OBJ).ImportDir("/a/b/c","abc.xml","/t2",,1)`))
		})
	})
})
