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
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	. "github.com/ontariosystems/isclib"
)

var _ = Describe("ToggleZSTU", func() {
	const (
		path  = "/test/cache/cache.cpf"
		zstu0 = "some line that isn't ZSTU\nanother line\nZSTU=0\nanother line\n"
		zstu1 = "some line that isn't ZSTU\nanother line\nZSTU=1\nanother line\n"
	)

	BeforeEach(func() {
		FS = new(afero.MemMapFs)
		FS.MkdirAll(filepath.Dir(path), 0755)
	})

	Context("with ZSTU=0", func() {
		BeforeEach(func() {
			Expect(afero.WriteFile(FS, path, []byte(zstu0), 0644)).To(Succeed())
		})

		Context("when toggling to true", func() {
			It("toggles ZSTU line to ZSTU=1", func() {
				Expect(ToggleZSTU(path, true)).To(BeFalse())
				Expect(afero.ReadFile(FS, path)).To(WithTransform(toStr, Equal(zstu1)))
			})
		})

		Context("when toggling to false", func() {
			It("toggles ZSTU line to ZSTU=0", func() {
				Expect(ToggleZSTU(path, false)).To(BeFalse())
				Expect(afero.ReadFile(FS, path)).To(WithTransform(toStr, Equal(zstu0)))
			})
		})
	})

	Context("with ZTU=1", func() {
		BeforeEach(func() {
			Expect(afero.WriteFile(FS, path, []byte(zstu1), 0644)).To(Succeed())
		})

		Context("when toggling to false", func() {
			It("toggles ZSTU line to ZSTU=0", func() {
				Expect(ToggleZSTU(path, false)).To(BeTrue())
				Expect(afero.ReadFile(FS, path)).To(WithTransform(toStr, Equal(zstu0)))
			})
		})

		Context("when toggling to true", func() {
			It("toggles ZSTU line to ZSTU=1", func() {
				Expect(ToggleZSTU(path, true)).To(BeTrue())
				Expect(afero.ReadFile(FS, path)).To(WithTransform(toStr, Equal(zstu1)))
			})
		})
	})
})

func toStr(b []byte) string {
	return string(b)
}
