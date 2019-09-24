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

func (c *Commands) Set(command Commands)     { *c |= command }
func (c *Commands) Clear(command Commands)   { *c &^= command }
func (c *Commands) Toggle(command Commands)  { *c ^= command }
func (c Commands) Has(command Commands) bool { return c&command != 0 }

const (
	CControlCommand Commands = 1 << iota
	CSessionCommand
	IrisCommand
	NoCommand Commands = 0
)

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
		log.WithField("csessionPath", globalCControlPath).WithError(err).Debug("csession executable not found")
	}

	return commands
}
