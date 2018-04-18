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

// An InstanceStatus represents one of the various status associated with Caché/Ensemble instances.
type InstanceStatus string

const (
	// InstanceStatusUnknown represents a blank/unknown instance status.
	InstanceStatusUnknown InstanceStatus = ""

	// InstanceStatusRunning represents a running instance.
	InstanceStatusRunning InstanceStatus = "running"

	// InstanceStatusInhibited represents an instance that is up but sign-ons have been inhibited due to an issue.
	InstanceStatusInhibited InstanceStatus = "sign-on inhibited"

	// InstanceStatusPrimaryTransition represents an instance that is up but the primary mirror member is being determined.
	InstanceStatusPrimaryTransition InstanceStatus = "sign-on inhibited:primary transition"

	// InstanceStatusDown represents an instance that is down.
	InstanceStatusDown InstanceStatus = "down"

	// InstanceStatusMissingIDS represents an instance that is up but missing a non-critical (but expected) information file.
	InstanceStatusMissingIDS InstanceStatus = "running on node ? (cache.ids missing)"
)

// Handled will return true when this status is a known and handled status.
func (iis InstanceStatus) Handled() bool {
	switch iis {
	default:
		return false
	case
		InstanceStatusRunning,
		InstanceStatusInhibited,
		InstanceStatusPrimaryTransition,
		InstanceStatusDown,
		InstanceStatusMissingIDS:
		return true
	}
}

// Ready will return true if the status represents an acceptably running status.
func (iis InstanceStatus) Ready() bool {
	switch iis {
	default:
		return false
	case
		InstanceStatusRunning,
		InstanceStatusMissingIDS:
		return true
	}
}

// Up will return true if status represents any up status (even unclean states like sign-on inhibited)
func (iis InstanceStatus) Up() bool {
	switch iis {
	default:
		return false
	case
		InstanceStatusRunning,
		InstanceStatusInhibited,
		InstanceStatusPrimaryTransition,
		InstanceStatusMissingIDS:
		return true
	}
}

// Down will return true if the instance status represents a fully down instance.
func (iis InstanceStatus) Down() bool {
	switch iis {
	default:
		return false
	case
		InstanceStatusDown:
		return true
	}
}

// RequiresBypass returns true when a bypass is required to stop the instance
func (iis InstanceStatus) RequiresBypass() bool {
	switch iis {
	default:
		return false
	case
		InstanceStatusInhibited,
		InstanceStatusPrimaryTransition:
		return true
	}
}
