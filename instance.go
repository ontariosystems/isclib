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
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	irisKeyName             = "license.key"
	cacheKeyName            = "cache.key"
	primaryJournalPattern   = "CurrentDirectory=(.+)"
	alternateJournalPattern = "AlternateDirectory=(.+)"
	//regex to remove the [ ,1,,, etc. ] configuration on InterSystems DAT lines
	extraInfoPattern  = "(1+|,+)"
	managerUserKey    = "security_settings.manager_user"
	managerGroupKey   = "security_settings.manager_group"
	ownerUserKey      = "security_settings.cache_user"
	ownerGroupKey     = "security_settings.cache_group"
	irisOwnerUserKey  = "security_settings.iris_user"
	irisOwnerGroupKey = "security_settings.iris_group"
	// DefaultImportQualifiers are the default ISC qualifiers used for importing source
	DefaultImportQualifiers = "/compile/keepsource/expand/multicompile"
	// CacheDatName is the common name for a Cache database file
	CacheDatName = "CACHE.DAT"
	// IrisDatName is the common name for a Iris database file
	IrisDatName = "IRIS.DAT"
)

var (
	// ErrLoadFailed is an error signifying that the loading of the source code failed
	ErrLoadFailed   = errors.New("load did not appear to finish successfully")
	getQlist        = qlist
	parameterReader = fileParameterReader
)

// An Instance represents an instance of Caché/Ensemble/Iris on the current system.
type Instance struct {
	// Required to be able to run the executor
	SessionPath string `json:"-"` // The path to the session executable
	ControlPath string `json:"-"` // The path to the control executable

	// These values come directly from qlist
	Name             string         `json:"name"`             // The name of the instance
	Directory        string         `json:"directory"`        // The directory in which the instance is installed
	Version          string         `json:"version"`          // The version of Caché/Ensemble/Iris
	Status           InstanceStatus `json:"status"`           // The status of the instance (down, running, etc.)
	Activity         string         `json:"activity"`         // The last activity date and time (as a string)
	CPFFileName      string         `json:"cpfFileName"`      // The name of the CPF file used by this instance at startup
	SuperServerPort  int            `json:"superServerPort"`  // The SuperServer port
	WebServerPort    int            `json:"webServerPort"`    // The internal WebServer port
	JDBCPort         int            `json:"jdbcPort"`         // The JDBC port
	State            string         `json:"state"`            // The State of the instance (warn, etc.)
	Product          Product        `json:"product"`          // The product name of the instance
	MirrorMemberType string         `json:"mirrorMemberType"` // The mirror member type (Failover, Disaster Recovery, etc)
	MirrorStatus     string         `json:"mirrorStatus"`     // The mirror Status (Primary, Backup, Connected, etc.)
	DataDirectory    string         `json:"dataDirectory"`    //  The instance data directory.  This might be the same as Directory if durable %SYS isn't in use

	executionSysProcAttr *syscall.SysProcAttr // This is used internally to allow execution of Caché code as different users
}

// Update will query the the underlying instance and update the Instance fields with its current state.
// It returns any error encountered.
func (i *Instance) Update() error {
	procAttr, err := i.managerSysProc()
	if err != nil {
		return err
	}

	// if we didn't get a manager proc, try to update without it to find the manager
	if procAttr == nil {
		q, err := getQlist(i.Name, nil)
		if err != nil {
			return err
		}

		if err := i.UpdateFromQList(q); err != nil {
			return err
		}

		procAttr, err = i.managerSysProc()
		if err != nil {
			return err
		}
	}

	q, err := getQlist(i.Name, procAttr)
	if err != nil {
		return err
	}

	return i.UpdateFromQList(q)
}

// UpdateFromQList will update the current Instance with the values from the qlist string.
// It returns any error encountered.
func (i *Instance) UpdateFromQList(qlist string) (err error) {
	qs := strings.Split(qlist, "^")
	if len(qs) < 8 {
		return fmt.Errorf("insufficient pieces in qlist, need at least 8, qlist: %s", qlist)
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
	i.DataDirectory = i.Directory
	i.Version = qs[2]
	i.Status, i.Activity = qlistStatus(qs[3])
	i.CPFFileName = qs[4]
	if len(qs) == 8 {
		i.State = "ok"
	} else {
		i.State = qs[8]
	}

	var productString = ""
	if len(qs) >= 10 {
		// Changes to this could make this string misidentify the product
		// It could be that the value won't match any of the known product strings we check in which case you would have the product reported as Cache
		productString = qs[9]
	}
	i.Product = i.determineProduct(productString)

	if len(qs) >= 11 {
		i.MirrorMemberType = qs[10]
	}

	if len(qs) >= 12 {
		i.MirrorStatus = qs[11]
	}

	if len(qs) >= 13 && qs[12] != "" {
		i.DataDirectory = qs[12]
	}

	return nil
}

// Dat holds information that pertains an existing ISC database
type Dat struct {
	Path       string
	Permission string
	Owner      string
	Group      string
	Exists     bool
}

// DatInfo will parse the instance's CPF file for its databases (CACHE.DAT, IRIS.DAT).
// It will get the path of the InterSystems DAT file, the permissions on it, and its owning user / group.
// The function returns a map of Dat structs containing the above information using the name of the database as its key.
func (i *Instance) DatInfo() (map[string]Dat, error) {
	cpfPath := filepath.Join(i.DataDirectory, i.CPFFileName)
	file, err := os.Open(cpfPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var inDbSection bool
	var dats = make(map[string]Dat)
	re := regexp.MustCompile(extraInfoPattern)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := re.ReplaceAllString(scanner.Text(), "")

		if inDbSection {
			if strings.TrimSpace(line) == "" {
				break
			}
			splitLine := strings.Split(line, "=")
			iscDatPath := splitLine[1] + i.DetermineISCDatFileName()
			iscDat := Dat{Path: splitLine[1], Exists: true}
			datFileInfo, err := os.Stat(iscDatPath)
			if err != nil {
				if os.IsNotExist(err) {
					iscDat.Exists = false
				} else {
					return nil, err
				}
			} else {
				fileOwner, err := user.LookupId(fmt.Sprint(datFileInfo.Sys().(*syscall.Stat_t).Uid))
				if err != nil {
					return nil, err
				}
				iscDat.Owner = fileOwner.Username
				fileGroup, err := user.LookupGroupId(fmt.Sprint(datFileInfo.Sys().(*syscall.Stat_t).Gid))
				if err != nil {
					return nil, err
				}
				iscDat.Group = fileGroup.Name
				iscDat.Permission = datFileInfo.Mode().String()
			}
			dats[splitLine[0]] = iscDat
		} else if line == "[Databases]" {
			inDbSection = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return dats, nil
}

// DetermineManager will determine the manager of an instance by reading the parameters file associated with this instance.
// The manager is the primary user of the instance that will be able to perform start/stop operations etc.
// It returns the manager and manager group as strings and any error encountered.
func (i *Instance) DetermineManager() (string, string, error) {
	return i.getUserAndGroupFromParameters("Manager", managerUserKey, managerGroupKey)
}

// managerSysProc is used to run instance management commands as a different user (if the current user isn't the manager)
func (i *Instance) managerSysProc() (*syscall.SysProcAttr, error) {
	// can't find manager if we don't have a directory
	if i.Directory == "" {
		return nil, nil
	}

	mgr, _, err := i.DetermineManager()
	if err != nil {
		return nil, err
	}

	uid, gid, err := lookupUser(mgr)
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{"user": mgr, "uid": uid, "gid": gid}).Debug("instance manager sysproc")
	return &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: uint32(uid),
			Gid: uint32(gid),
		},
	}, nil
}

// DetermineOwner will determine the owner of an instance by reader the parameters file associate with this instance.
// The owner is the user which owns the files from the installers and as who most Caché processes will be running.
// It returns the owner and owner group as strings and any error encountered.
func (i *Instance) DetermineOwner() (string, string, error) {
	switch i.Product {
	case Iris:
		return i.getUserAndGroupFromParameters("Owner", irisOwnerUserKey, irisOwnerGroupKey)
	default:
		return i.getUserAndGroupFromParameters("Owner", ownerUserKey, ownerGroupKey)
	}
}

// DeterminePrimaryJournalDirectory will parse the ISC instance's CPF file for its primary journal directory (CurrentDirectory).
func (i *Instance) DeterminePrimaryJournalDirectory() (string, error) {
	cpfPath := filepath.Join(i.DataDirectory, i.CPFFileName)
	file, err := os.Open(cpfPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	re, err := regexp.Compile(primaryJournalPattern)
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		matches := re.FindStringSubmatch(scanner.Text())
		if len(matches) > 0 {
			return matches[1], nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("unable to determine primary journal directory")
}

// DetermineSecondaryJournalDirectory will parse the ISC instance's CPF file for its secondary journal directory (AlternateDirectory).
func (i *Instance) DetermineSecondaryJournalDirectory() (string, error) {
	cpfPath := filepath.Join(i.DataDirectory, i.CPFFileName)
	file, err := os.Open(cpfPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	re, err := regexp.Compile(alternateJournalPattern)
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		matches := re.FindStringSubmatch(scanner.Text())
		if len(matches) > 0 {
			return matches[1], nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("unable to determine secondary journal directory")
}

// DetermineISCDatFileName returns the filename of the InterSystems DAT files used by the instance
func (i *Instance) DetermineISCDatFileName() string {
	switch i.Product {
	case Iris:
		return IrisDatName
	default:
		return CacheDatName
	}
}

// LicenseKeyFilePath returns the file path to the license key for the instance
func (i *Instance) LicenseKeyFilePath() string {
	switch i.Product {
	case Iris:
		return filepath.Join(i.DataDirectory, "mgr", irisKeyName)
	default:
		return filepath.Join(i.DataDirectory, "mgr", cacheKeyName)
	}
}

// Start will ensure that an instance is started.
// It returns any error encountered when attempting to start the instance.
func (i *Instance) Start() error {
	// TODO: Think about a nozstu flag if there's a reason
	if i.Status.Down() {
		cmd := exec.Command(i.controlPath(), "start", i.Name, "quietly")
		procAttr, err := i.managerSysProc()
		if err != nil {
			return err
		}

		cmd.SysProcAttr = procAttr
		if output, err := cmd.CombinedOutput(); err != nil {
			log.WithError(err).WithFields(log.Fields{"output": string(output), "instance": i.Name}).Debug("Error start quietly")
			return fmt.Errorf("error starting instance, error: %w", err)
		}
	}

	if err := i.Update(); err != nil {
		return fmt.Errorf("error refreshing instance state during start, error: %w", err)
	}

	if !i.Status.Ready() {
		return fmt.Errorf("failed to start instance, name: %s, status: %s", i.Name, i.Status)
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
		cmd := exec.Command(i.controlPath(), args...)
		procAttr, err := i.managerSysProc()
		if err != nil {
			return err
		}
		cmd.SysProcAttr = procAttr
		if output, err := cmd.CombinedOutput(); err != nil {
			ilog.WithError(err).WithFields(log.Fields{"output": string(output), "args": args}).Debug("Error stopping")
			return fmt.Errorf("error stopping instance, error: %w", err)
		}
	}

	if err := i.Update(); err != nil {
		return fmt.Errorf("error refreshing instance state during stop, error: %w", err)
	}

	if !i.Status.Down() {
		return fmt.Errorf("failed to stop instance, name: %s, status: %s", i.Name, i.Status)
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

// ExecuteAsManager will configure the instance to execute all future commands as the instance's owner.
// This command only functions if the calling program is running as root.
// It returns any error encountered.
func (i *Instance) ExecuteAsManager() error {
	owner, _, err := i.DetermineManager()
	if err != nil {
		return err
	}

	return i.ExecuteAsUser(owner)
}

// ExecuteAsUser will configure the instance to execute all future commands as the provided user.
// This command only functions if the calling program is running as root.
// It returns any error encountered.
func (i *Instance) ExecuteAsUser(execUser string) error {
	// no need to switch users if we're already who we want to be
	if err := checkUser(execUser); err == nil {
		return nil
	}

	if err := checkUser("root"); err != nil {
		return err
	}

	uid, gid, err := lookupUser(execUser)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{"user": execUser, "uid": uid, "gid": gid}).Debug("Configured to execute as alternate user")
	i.executionSysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: uint32(uid),
			Gid: uint32(gid),
		},
	}
	return nil
}

func lookupUser(execUser string) (uid, gid uint64, err error) {
	var u *user.User
	u, err = user.Lookup(execUser)
	if err != nil {
		return
	}

	uid, err = strconv.ParseUint(u.Uid, 10, 32)
	if err != nil {
		return
	}

	gid, err = strconv.ParseUint(u.Gid, 10, 32)

	return
}

// ImportSource will import the source specified using a glob pattern into Caché with the provided qualifiers.
// sourcePathGlob only allows a subset of glob patterns.  It must be in the format /p/a/t/h/**/*.xml
//
//	/p/a/t/h/ is the import directory
//	you have have at most one **
//	after the ** you must have only a file pattern
//	To import a single file it would be /a/b/c/file.xml
//
// qualifiers are standard Caché import/compile qualifiers, if none are provided a default set will be used
// It returns any output of the import and any error encountered.
func (i *Instance) ImportSource(namespace, sourcePathGlob string, qualifiers ...string) (string, error) {
	qstr := strings.TrimSpace(strings.Join(qualifiers, ""))
	if qstr == "" {
		qstr = DefaultImportQualifiers
	}

	id, err := NewImportDescription(sourcePathGlob, qstr)
	if err != nil {
		return "", err
	}

	cmd := id.String()
	l := log.WithFields(log.Fields{
		"instance":   i.Name,
		"namespace":  namespace,
		"path":       sourcePathGlob,
		"qualifiers": qstr,
		"command":    cmd,
	})
	l.Debug("Attempting to import source")
	o, err := i.SessionCommand(namespace, cmd).CombinedOutput()
	out := string(o)
	l.WithField("output", out).Debug("import command result")
	if err != nil {
		return out, err
	}

	if !strings.Contains(out, "Load finished successfully.") {
		return out, ErrLoadFailed
	}

	return out, nil
}

// Execute will read code from the provided io.Reader and execute it in the provided namespace.
// The code must be valid Caché ObjectScript INT code obeying all of the correct spacing with a MAIN label as the primary entry point.
// Valid INT code means (this list is not exhaustive)...
//   - Labels start at the first character on the line
//   - Non-labels start with a single space
//   - You may not have blank lines internal to the code
//   - You must have a single blank line at the end of the script
//
// It returns any output of the execution and any error encountered.
func (i *Instance) Execute(namespace string, codeReader io.Reader) (string, error) {
	var out bytes.Buffer
	err := i.ExecuteWithOutput(namespace, codeReader, &out)
	return out.String(), err
}

// ExecuteWithOutput will read code from the provided io.Reader and execute it in the provided namespace while
// writing any output to the provided io.Writer.
func (i *Instance) ExecuteWithOutput(namespace string, codeReader io.Reader, out io.Writer) error {
	elog := log.WithField("namespace", namespace)
	elog.Debug("Attempting to execute INT code")

	codePath, err := i.genExecutorTmpFile(codeReader)
	if err != nil {
		return err
	}
	elog.WithField("path", codePath).Debug("Acquired temporary file")

	defer os.Remove(codePath)

	if output, err := i.ImportSource(namespace, codePath, "/compile", "/keepsource"); err != nil {
		elog.WithError(err).WithField("output", output).Error("unable to import")
		return err
	}

	routineName := filepath.Base(codePath)
	defer func() {
		if err := i.removeTempRoutine(namespace, routineName); err != nil {
			log.WithError(err).Error("Failed to remove temp routine")
		}
	}()

	cmd := i.SessionCommand(namespace, "EnsLibMain^"+routineName)

	cmd.Stdout = out
	if err := cmd.Start(); err != nil {
		log.WithError(err).Debug("Failed to start session")
		return err
	}

	elog.Debug("Waiting on session to exit")
	return cmd.Wait()
}

// SessionCommand will return a properly configured instance of exec.Cmd to
// run the provided command (properly formatted for session) in the provided
// namespace.
func (i *Instance) SessionCommand(namespace, command string) *exec.Cmd {
	args := []string{i.Name}
	if namespace != "" {
		args = append(args, "-U", namespace)
	}

	if command != "" {
		args = append(args, command)
	}

	sc := i.sessionCommand()
	if strings.Contains(sc, " ") {
		scp := strings.Split(sc, " ")
		sc = scp[0]
		args = append(scp[1:], args...)
	}
	log.WithFields(log.Fields{"instance": i.Name, "cmd": sc, "args": args}).Debug("session command")
	cmd := exec.Command(sc, args...)
	if i.executionSysProcAttr != nil {
		cmd.SysProcAttr = i.executionSysProcAttr
	}

	return cmd
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
	f, err := parameterReader(i.Directory, iscParametersFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return LoadParametersISC(f)
}

func fileParameterReader(directory string, file string) (io.ReadCloser, error) {
	pfp := filepath.Join(directory, file)
	f, err := os.Open(pfp)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// WaitForReady waits indefinitely for an instance to be up and ready for use
func (i *Instance) WaitForReady(ctx context.Context) error {
	return i.WaitForReadyWithInterval(ctx, 100*time.Millisecond)
}

// WaitForReadyWithInterval waits for an instance to be up and ready for use or until the interval is exceeded
func (i *Instance) WaitForReadyWithInterval(ctx context.Context, interval time.Duration) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
			_ = i.Update()
			if i.Status.Ready() {
				return nil
			}
		}
	}
}

func qlistStatus(statusAndTime string) (InstanceStatus, string) {
	s := strings.SplitN(statusAndTime, ",", 2)
	var a string
	if len(s) > 1 {
		a = strings.TrimSpace(s[1])
	}
	return InstanceStatus(strings.ToLower(s[0])), a
}

func (i *Instance) genExecutorTmpFile(codeReader io.Reader) (path string, error error) {
	tmpFile, err := os.CreateTemp(executeTemporaryDirectory, "ELEXEC")
	if err != nil {
		return "", err
	}

	defer tmpFile.Close()

	if err := os.Chmod(tmpFile.Name(), 0644); err != nil {
		return "", fmt.Errorf("failed to set permissions on import file: %w", err)
	}

	routineName := filepath.Base(tmpFile.Name())
	if _, err := tmpFile.Write([]byte(fmt.Sprintf(importXMLHeader, routineName))); err != nil {
		return "", fmt.Errorf("failed to write XML header: %w", err)
	}

	if _, err := io.Copy(tmpFile, codeReader); err != nil {
		return "", err
	}

	if _, err := tmpFile.Write([]byte(importXMLFooter)); err != nil {
		return "", fmt.Errorf("failed to write XML footer: %w", err)
	}

	// Need to set the permissions here or the file will be owned by root and the execution will fail
	if i.executionSysProcAttr != nil {
		if err := os.Chown(
			tmpFile.Name(),
			int(i.executionSysProcAttr.Credential.Uid),
			int(i.executionSysProcAttr.Credential.Gid),
		); err != nil {
			return "", fmt.Errorf("failed to set ownership on import file: %w", err)
		}
	}

	return tmpFile.Name(), nil
}

func (i *Instance) sessionCommand() string {
	if i.SessionPath == "" {
		switch i.Product {
		case Iris:
			return globalIrisSessionCommand
		default:
			return globalCSessionPath
		}
	}

	return i.SessionPath
}

func (i *Instance) controlPath() string {
	if i.ControlPath == "" {
		switch i.Product {
		case Iris:
			return globalIrisPath
		default:
			return globalCControlPath
		}
	}

	return i.ControlPath
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

func (i *Instance) removeTempRoutine(namespace, path string) error {
	routineName := filepath.Base(path)
	l := log.WithFields(log.Fields{
		"instance":  i.Name,
		"namespace": namespace,
		"routine":   routineName,
	})

	l.Debug("Removing temporary routine")
	cmd := i.SessionCommand(namespace, fmt.Sprintf(`##class(%%Routine).Delete("%s",0,1)`, routineName))
	if err := cmd.Start(); err != nil {
		l.WithError(err).Error("Failed to start deletion")
		return fmt.Errorf("failed to start routine deletion: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		l.WithError(err).Error("Failed to delete routine")
		return fmt.Errorf("failed to execute routine deletion: %w", err)
	}

	return nil
}

func (i *Instance) determineProduct(product string) Product {
	if product != "" {
		return ParseProduct(product)
	}

	pi, err := i.ReadParametersISC()
	if err != nil {
		return Cache
	}

	return ParseProduct(pi.Value("product_info.name"))
}
