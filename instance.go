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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
)

type Instance struct {
	// Required to be able to run the executor
	CSessionPath string `json:"-"`
	CControlPath string `json:"-"`

	// These values come directly from ccontrol qlist
	Name            string         `json:"name"`
	Directory       string         `json:"directory"`
	Version         string         `json:"version"`
	Status          InstanceStatus `json:"status"`
	Activity        string         `json:"activity"`
	CPFFileName     string         `json:"cpfFileName"`
	SuperServerPort int            `json:"superServerPort"`
	WebServerPort   int            `json:"webServerPort"`
	JDBCPort        int            `json:"jdbcPort"`
	State           string         `json:"state"`
	// There appears to be an additional property after state but I don't know what it is!

	// Values not from ccontrol qlist
	owner string `json:"owner"`
}

func (i *Instance) Update() error {
	q, err := qlist(i.Name)
	if err != nil {
		return err
	}

	return i.UpdateFromQList(q)
}

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

func (i *Instance) DetermineOwner() (string, error) {
	pfp := filepath.Join(i.Directory, cacheParametersFile)
	f, err := os.Open(pfp)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var owner string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		m := cacheOwnerRegex.FindStringSubmatch(scanner.Text())
		if len(m) == 2 {
			owner = m[1]
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	if owner == "" {
		return "", fmt.Errorf("Failed to find cache owner in %s", cacheParametersFile)
	}

	return owner, nil
}

// TODO: Think about a nozstu flag if there's a reason
func (i *Instance) Start() error {
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

// This will execute the label MAIN from the properly formatted Cache INT code stored in the codeReader in namespace
func (i *Instance) Execute(namespace string, codeReader io.Reader) (output string, err error) {
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
		return "", nil
	}

	return tmpFile.Name(), nil
}

func (i *Instance) csessionPath() string {
	if i.CSessionPath == "" {
		return defaultCSessionPath
	}

	return i.CSessionPath
}

func (i *Instance) ccontrolPath() string {
	if i.CControlPath == "" {
		return defaultCControlPath
	}

	return i.CControlPath
}
