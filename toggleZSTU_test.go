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
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	. "github.com/ontariosystems/isclib"
)

var _ = Describe("ToggleZSTU", func() {
	var dir = "/test/cache"
	var name = "cache.cpf"
	var path = filepath.Join(dir, name)
	var content = `
	some line that isn't ZSTU
	another line
	ZSTU=1
	another line
	`
	BeforeEach(func() {
		FS = new(afero.MemMapFs)
		FS.MkdirAll(dir, 0755)
		afero.WriteFile(FS, path, []byte(content), 0644)
	})

	Describe("Open cpf file for reading", func() {
		It("toggles ZSTU line to ZSTU=0", func() {
			err := ToggleZSTU(path, false)
			Expect(err).NotTo(HaveOccurred())
		})

	})

})
