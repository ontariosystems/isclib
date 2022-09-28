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
	"fmt"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

// qlist returns the results of executing qlist for the specified instance.
// If instanceName is "", it will return the results of an argumentless qlist (which contains all instances)
// It returns a string containing the combined standard input and output of the qlist command and any error which occurred.
func qlist(instanceName string) (string, error) {
	// Example qlist output...
	// DOCKER^/ensemble/instances/docker/^2015.2.2.805.0.16216^down, last used Fri May 13 18:12:33 2016^cache.cpf^56772^57772^62972^^
	// DOCKER^/ensemble/instances/docker^2018.1.1.643.0^running, since Mon Jul 23 14:42:09 2018^iris.cpf^1972^57772^62972^ok^IRIS^^^/ensemble/instances/docker
	qlist := ""
	args := []string{"qlist"}
	if instanceName != "" {
		args = append(args, instanceName)
	}

	var cmd *exec.Cmd
	commands := AvailableCommands()
	switch {
	case commands.Has(IrisCommand):
		cmd = exec.Command(globalIrisPath, args...)
	case commands.Has(CControlCommand):
		cmd = exec.Command(globalCControlPath, args...)
	default:
		return qlist, nil
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"output": string(out), "command": cmd.Path, "args": cmd.Args}).Debug("Error running qlist")
		return "", fmt.Errorf("error running qlist: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}
