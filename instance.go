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
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
)

const (
	managerUserKey  = "security_settings.manager_user"
	managerGroupKey = "security_settings.manager_group"
	ownerUserKey    = "security_settings.cache_user"
	ownerGroupKey   = "security_settings.cache_group"
)

// An Instance represents an instance of Caché/Ensemble on the current system.
type Instance struct {
	// Required to be able to run the executor
	CSessionPath string `json:"-"` // The path to the csession executable
	CControlPath string `json:"-"` // The path to the ccontrol executable

	// These values come directly from ccontrol qlist
	Name            string         `json:"name"`            // The name of the instance
	Directory       string         `json:"directory"`       // The directory in which the instance is installed
	Version         string         `json:"version"`         // The version of Caché/Ensemble
	Status          InstanceStatus `json:"status"`          // The status of the instance (down, running, etc.)
	Activity        string         `json:"activity"`        // The last activity date and time (as a string)
	CPFFileName     string         `json:"cpfFileName"`     // The name of the CPF file used by this instance at startup
	SuperServerPort int            `json:"superServerPort"` // The SuperServer port
	WebServerPort   int            `json:"webServerPort"`   // The internal WebServer port
	JDBCPort        int            `json:"jdbcPort"`        // The JDBC port
	State           string         `json:"state"`           // The State of the instance (warn, etc.)
	// There appears to be an additional property after state but I don't know what it is!

	executionSysProcAttr *syscall.SysProcAttr // This is used internally to allow execution of Caché code as different users
}

// Update will query the the underlying instance and update the Instance fields with its current state.
// It returns any error encountered.
func (i *Instance) Update() error {
	q, err := qlist(i.Name)
	if err != nil {
		return err
	}

	return i.UpdateFromQList(q)
}

// UpdateFromQList will update the current Instance with the values from the qlist string.
// It returns any error encountered.
func (i *Instance) UpdateFromQList(qlist string) (err error) {
	qs := strings.Split(qlist, "^")
	if len(qs) < 9 {
		return fmt.Errorf("Insufficient pieces in qlist, need at least 9, qlist: %s", qlist)
	}

	if i.SuperServerPort, err = strconv.Atoi(qs[5]); err != nil {
		return err
	}

	if i.WebServerPort, err = strconv.Atoi(qs[6]); err != nil {
		return err
	}

	if i.JDBCPort, err = strconv.Atoi(qs[7]); err != nil {
		return err
	}

	i.Name = qs[0]
	i.Directory = qs[1]
	i.Version = qs[2]
	i.Status, i.Activity = qlistStatus(qs[3])
	i.CPFFileName = qs[4]
	i.State = qs[8]

	return nil
}

// DetermineManager will determine the manager of an instance by reading the parameters file associated with this instance.
// The manager is the primary user of the instance that will be able to perform start/stop operations etc.
// It returns the manager and manager group as strings and any error encountered.
func (i *Instance) DetermineManager() (string, string, error) {
	return i.getUserAndGroupFromParameters("Manager", managerUserKey, managerGroupKey)
}

// DetermineOwner will determine the owner of an instance by reader the parameters file associate with this instance.
// The owner is the user which owns the files from the installers and as who most Caché processes will be running.
// It returns the owner and owner group as strings and any error encountered.
func (i *Instance) DetermineOwner() (string, string, error) {
	return i.getUserAndGroupFromParameters("Owner", ownerUserKey, ownerGroupKey)
}

// Start will ensure that an instance is started.
// It returns any error encountered when attempting to start the instance.
func (i *Instance) Start() error {
	// TODO: Think about a nozstu flag if there's a reason
	if i.Status.Down() {
		if output, err := exec.Command(i.ccontrolPath(), "start", i.Name, "quietly").CombinedOutput(); err != nil {
			return fmt.Errorf("Error starting instance, error: %s, output: %s", err, output)
		}
	}

	if err := i.Update(); err != nil {
		return fmt.Errorf("Error refreshing instance state during start, error: %s", err)
	}

	if !i.Status.Ready() {
		return fmt.Errorf("Failed to start instance, name: %s, status: %s", i.Name, i.Status)
	}

	return nil
}

// Stop will ensure that an instance is started.
// It returns any error encountered when attempting to stop the instance.
func (i *Instance) Stop() error {
	ilog := log.WithField("name", i.Name)
	ilog.Debug("Shutting down instance")
	if i.Status.Up() {
		args := []string{"stop", i.Name}
		if i.Status.RequiresBypass() {
			args = append(args, "bypass")
		}
		args = append(args, "quietly")
		if output, err := exec.Command(i.ccontrolPath(), args...).CombinedOutput(); err != nil {
			return fmt.Errorf("Error stopping instance, error: %s, output: %s", err, output)
		}
	}

	if err := i.Update(); err != nil {
		return fmt.Errorf("Error refreshing instance state during stop, error: %s", err)
	}

	if !i.Status.Down() {
		return fmt.Errorf("Failed to stop instance, name: %s, status: %s", i.Name, i.Status)
	}

	return nil
}

// ExecuteAsCurrentUser will configure the instance to execute all future commands as the current user.
// It returns any error encountered.
func (i *Instance) ExecuteAsCurrentUser() error {
	log.Debug("Removing execution user")
	i.executionSysProcAttr = nil
	return nil
}

// ExecuteAsManager will  configure the instance to execute all future commands as the instance's owner.
// This command only functions if the calling program is running as root.
// It returns any error encountered.
func (i *Instance) ExecuteAsManager() error {
	owner, _, err := i.DetermineManager()
	if err != nil {
		return err
	}

	return i.ExecuteAsUser(owner)
}

// ExecuteAsUser will  configure the instance to execute all future commands as the provided user.
// This command only functions if the calling program is running as root.
// It returns any error encountered.
func (i *Instance) ExecuteAsUser(execUser string) error {
	// TODO: Check for euid 0 instead of just letting it fail in an arbitrary function
	u, err := user.Lookup(execUser)
	if err != nil {
		return err
	}

	uid, err := strconv.ParseUint(u.Uid, 10, 32)
	if err != nil {
		return err
	}

	gid, err := strconv.ParseUint(u.Gid, 10, 32)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{"user": execUser, "uid": u.Uid, "gid": u.Gid}).Debug("Configured to execute as alternate user")
	i.executionSysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: uint32(uid),
			Gid: uint32(gid),
		},
	}
	return nil
}

// Execute will read code from the provided io.Reader ane execute it in the provided namespace.
// The code must be valid Caché ObjectScript INT code obeying all of the correct spacing with a MAIN label as the primary entry point.
// Valid INT code means (this list is not exhaustive)...
//   - Labels start at the first character on the line
//   - Non-labels start with a single space
//   - You may not have blank lines internal to the code
//   - You must have a single blank line at the end of the script
// It returns any output of the execution and any error encountered.
func (i *Instance) Execute(namespace string, codeReader io.Reader) (string, error) {
	// TODO: Async standard out from csession
	elog := log.WithField("namespace", namespace)
	elog.Debug("Attempting to execute INT code")

	codePath, err := i.genExecutorTmpFile(codeReader)
	if err != nil {
		return "", err
	}
	elog.WithField("path", codePath).Debug("Acquired temporary file")

	defer os.Remove(codePath)

	// Not using -U because it won't work if the user has a startup namespace
	cmd := exec.Command(i.csessionPath(), i.Name)

	if i.executionSysProcAttr != nil {
		cmd.SysProcAttr = i.executionSysProcAttr
	}

	in, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Start(); err != nil {
		log.WithError(err).Debug("Failed to start csession")
		return "", err
	}

	importString := fmt.Sprintf(codeImportString, namespace, codePath)
	// TODO: Consider parsing the code and correcting indenting/spacing/etc
	elog.WithField("importCode", importString).Debug("Attempting to load INT code into buffer")
	if count, err := in.Write([]byte(importString)); err != nil {
		log.WithError(err).WithField("count", count).Debug("Attempted to write to csession stdin and failed")
		return "", err
	}
	// TODO: Add the required blank line at the end of the int code if it is missing
	in.Close()

	elog.Debug("Waiting on csession to exit")
	err = cmd.Wait()
	return out.String(), err
}

// ExecuteString will execute the provided code in the specified namespace.
// code must be properly formatted INT code. See the documentation for Execute for more information.
// It returns any output of the execution and any error encountered.
func (i *Instance) ExecuteString(namespace string, code string) (string, error) {
	b := bytes.NewReader([]byte(code))
	return i.Execute(namespace, b)
}

// ReadParametersISC will read the current instances parameters ISC file into a simple data structure.
// It returns the ParametersISC data structure and any error encountered.
func (i *Instance) ReadParametersISC() (ParametersISC, error) {
	pfp := filepath.Join(i.Directory, cacheParametersFile)
	f, err := os.Open(pfp)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return LoadParametersISC(f)
}

func qlistStatus(statusAndTime string) (InstanceStatus, string) {
	s := strings.SplitN(statusAndTime, ",", 2)
	var a string
	if len(s) > 1 {
		a = strings.TrimSpace(s[1])
	}
	return InstanceStatus(strings.ToLower(s[0])), a
}

func (i *Instance) genExecutorTmpFile(codeReader io.Reader) (string, error) {
	tmpFile, err := ioutil.TempFile("", "isclib-exec-")
	if err != nil {
		return "", err
	}

	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, codeReader); err != nil {
		return "", err
	}

	// Need to set the permissions here or the file will be owned by root and the execution will fail
	if i.executionSysProcAttr != nil {
		if err := os.Chown(
			tmpFile.Name(),
			int(i.executionSysProcAttr.Credential.Uid),
			int(i.executionSysProcAttr.Credential.Gid),
		); err != nil {
			return "", err
		}
	}

	return tmpFile.Name(), nil
}

func (i *Instance) csessionPath() string {
	if i.CSessionPath == "" {
		return globalCSessionPath
	}

	return i.CSessionPath
}

func (i *Instance) ccontrolPath() string {
	if i.CControlPath == "" {
		return globalCControlPath
	}

	return i.CControlPath
}

func (i *Instance) getUserAndGroupFromParameters(desc, userKey, groupKey string) (string, string, error) {
	pi, err := i.ReadParametersISC()
	if err != nil {
		return "", "", err
	}

	owner := pi.Value(userKey)
	if owner == "" {
		return "", "", fmt.Errorf("%s user not found in parameters file", desc)
	}

	group := pi.Value(groupKey)
	if group == "" {
		return "", "", fmt.Errorf("%s group not found in parameters file", desc)
	}

	return owner, group, nil
}
