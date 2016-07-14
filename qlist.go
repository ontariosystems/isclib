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
	"os/exec"
	"strings"
)

// qlist returns the results of executing ccontrol qlist for the specified instance.
// If instanceName is "", it will return the results of an argumentless qlist (which contains all instances)
// It returns a string containing the combined standard input and output of the qlist command and any error which occurred.
func qlist(instanceName string) (string, error) {
	// Example qlist output...
	// DOCKER^/ensemble/instances/docker/^2015.2.2.805.0.16216^down, last used Fri May 13 18:12:33 2016^cache.cpf^56772^57772^62972^^
	args := []string{"qlist"}
	if instanceName != "" {
		args = append(args, instanceName)
	}

	// TODO: Allow ccontrol path to be set on the package and make this use that value
	out, err := exec.Command(defaultCControlPath, args...).CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}
