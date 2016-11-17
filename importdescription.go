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
	TooManyRecursiveDirsErr   = errors.New("The glob must contain at most one **")
	MissingPathSeparatorErr   = errors.New("There must be a path separator between ** and file pattern")
	PathAfterRecursiveDirsErr = errors.New("A ** must only be used as the last portion of the path before the file pattern")
	WildcardInDirectoryErr    = errors.New("The directory portion of the glob must not contain *")
)

type ImportDescription struct {
	Dir         string
	FilePattern string
	Recursive   bool
	Qualifiers  string
}

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
			return nil, MissingPathSeparatorErr
		}

		if filepath.Dir(s[1]) != "/" {
			return nil, PathAfterRecursiveDirsErr
		}
		glob.FilePattern = filepath.Base(s[1])
		glob.Recursive = true
	default:
		return nil, TooManyRecursiveDirsErr
	}

	if strings.Contains(glob.Dir, "*") {
		return nil, WildcardInDirectoryErr
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

func (i *ImportDescription) String() string {
	var rec uint16
	if i.Recursive {
		rec = 1
	} else {
		rec = 0
	}
	return fmt.Sprintf(importFmtStr, i.Dir, i.FilePattern, i.Qualifiers, rec)
}
