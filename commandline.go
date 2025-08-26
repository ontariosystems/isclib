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
	"os/exec"

	log "github.com/sirupsen/logrus"
)

// Commands represents the ISC command lines that are available
type Commands uint

// Set updates the list of available commands to include the provided command
func (c *Commands) Set(command Commands) { *c |= command }

// Clear updates the list of available commands to remove the provided command
func (c *Commands) Clear(command Commands) { *c &^= command }

// Has returns a boolean indicating whether the requested command is marked as available
func (c Commands) Has(command Commands) bool { return c&command != 0 }

const (
	// CControlCommand indicates that the ccontrol command is available
	CControlCommand Commands = 1 << iota
	// CSessionCommand indicates that the csession command is available
	CSessionCommand
	// IrisCommand indicates that the iris command is available
	IrisCommand
	// NoCommand indicates the none of the ISC command lines are available
	NoCommand Commands = 0
)

// AvailableCommands returns a Commands bitmask indicating which ISC command lines are available
func AvailableCommands() Commands {
	var commands = NoCommand

	if _, err := exec.LookPath(globalIrisPath); err == nil {
		commands.Set(IrisCommand)
	} else {
		log.WithField("irisPath", globalIrisPath).WithError(err).Debug("iris executable not found")
	}

	if _, err := exec.LookPath(globalCControlPath); err == nil {
		commands.Set(CControlCommand)
	} else {
		log.WithField("controlPath", globalCControlPath).WithError(err).Debug("ccontrol executable not found")
	}

	if _, err := exec.LookPath(globalCSessionPath); err == nil {
		commands.Set(CSessionCommand)
	} else {
		log.WithField("csessionPath", globalCSessionPath).WithError(err).Debug("csession executable not found")
	}

	return commands
}
