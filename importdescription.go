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

package isclib

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	importFmtStr = `##class(%%SYSTEM.OBJ).ImportDir("%s","%s","%s",,%d)`
)

var (
	// ErrTooManyRecursiveDirs is an error signifying too many ** are included in the glob pattern
	ErrTooManyRecursiveDirs = errors.New("the glob must contain at most one **")
	// ErrMissingPathSeparator is an error signifying a missing path separator in the glob pattern
	ErrMissingPathSeparator = errors.New("there must be a path separator between ** and file pattern")
	// ErrPathAfterRecursiveDirs is an error signifying that additional path information is included after the ** in the glob pattern
	ErrPathAfterRecursiveDirs = errors.New("a ** must only be used as the last portion of the path before the file pattern")
	// ErrWildcardInDirectory is an error signifying that a wildcard has been included in the directory part of the glob pattern
	ErrWildcardInDirectory = errors.New("the directory portion of the glob must not contain *")
)

// ImportDescription holds information needed for constructing a valid ISC $SYSTEM.OBJ.ImportDir command
type ImportDescription struct {
	Dir         string
	FilePattern string
	Recursive   bool
	Qualifiers  string
}

// NewImportDescription creates and returns a new import description based on the provided glob pattern and ISC qualifiers
func NewImportDescription(pathGlob string, qualifiers string) (*ImportDescription, error) {
	glob := &ImportDescription{Qualifiers: qualifiers}

	s := strings.Split(pathGlob, "**")
	switch len(s) {
	case 1:
		glob.Dir = filepath.Dir(s[0])
		glob.FilePattern = filepath.Base(s[0])
	case 2:
		glob.Dir = filepath.Clean(s[0])
		if !strings.HasPrefix(s[1], "/") {
			return nil, ErrMissingPathSeparator
		}

		if filepath.Dir(s[1]) != "/" {
			return nil, ErrPathAfterRecursiveDirs
		}
		glob.FilePattern = filepath.Base(s[1])
		glob.Recursive = true
	default:
		return nil, ErrTooManyRecursiveDirs
	}

	if strings.Contains(glob.Dir, "*") {
		return nil, ErrWildcardInDirectory
	}

	if glob.Dir == "." {
		if cwd, err := os.Getwd(); err == nil {
			glob.Dir = cwd
		} else {
			return nil, err
		}
	}

	return glob, nil
}

// String returns an ISC $SYSTEM.OBJ.ImportDir command as a string
func (i *ImportDescription) String() string {
	var rec uint16
	if i.Recursive {
		rec = 1
	} else {
		rec = 0
	}
	return fmt.Sprintf(importFmtStr, i.Dir, i.FilePattern, i.Qualifiers, rec)
}
