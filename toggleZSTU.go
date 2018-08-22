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
	"bufio"
	"io"

	"github.com/spf13/afero"
)

// FS is a wrapper for the file system
var FS = afero.NewOsFs()

// ToggleZSTU ensures that the cpf file at the path provided has the ZSTU setting
// set to true or false based on the provided boolean value.  It also returns the
// original value for the ZSTU
func ToggleZSTU(cpfFilePath string, onOrOff bool) (originalValue bool, err error) {
	cpfFile, err := FS.Open(cpfFilePath)
	if err != nil {
		return originalValue, err
	}

	tmpFile, err := afero.TempFile(FS, "", "cpftemp")
	if err != nil {
		return originalValue, err
	}

	originalValue, err = parseAndWriteCPF(cpfFile, tmpFile, onOrOff)
	if err != nil {
		return originalValue, err
	}

	if err = cpfFile.Close(); err != nil {
		return originalValue, err
	}

	if err = tmpFile.Close(); err != nil {
		return originalValue, err
	}

	cpfFile, err = FS.Create(cpfFilePath)
	if err != nil {
		return originalValue, err
	}

	newCpfFile, err := FS.Open(tmpFile.Name())
	if err != nil {
		return originalValue, err
	}

	if _, err = io.Copy(cpfFile, newCpfFile); err != nil {
		return originalValue, err
	}

	if err := cpfFile.Close(); err != nil {
		return originalValue, err
	}

	if err := newCpfFile.Close(); err != nil {
		return originalValue, err
	}

	FS.Remove(tmpFile.Name())

	return originalValue, nil
}

func parseAndWriteCPF(cpfFile io.Reader, tmpFile io.Writer, onOrOff bool) (originalValue bool, err error) {
	scanner := bufio.NewScanner(cpfFile)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "ZSTU=1" || line == "ZSTU=0" {
			if line == "ZSTU=1" {
				originalValue = true
			} else {
				originalValue = false
			}
			if onOrOff {
				io.WriteString(tmpFile, "ZSTU=1\n")
			} else {
				io.WriteString(tmpFile, "ZSTU=0\n")
			}
		} else {
			io.WriteString(tmpFile, scanner.Text()+"\n")
		}
	}
	if err := scanner.Err(); err != nil {
		return originalValue, err
	}

	return originalValue, nil
}
