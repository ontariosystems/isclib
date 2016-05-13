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

type InstanceStatus string

const (
	InstanceStatusUnknown           InstanceStatus = ""
	InstanceStatusRunning           InstanceStatus = "running"
	InstanceStatusInhibited         InstanceStatus = "sign-on inhibited"
	InstanceStatusPrimaryTransition InstanceStatus = "sign-on inhibited:primary transition"
	InstanceStatusDown              InstanceStatus = "down"
	InstanceStatusMissingIDS        InstanceStatus = "running on node ? (cache.ids missing)"
)

// Returns true when the status is known and can be handled by this code
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

// Returns true if Cache is up and in a normal(ish) state
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

// Returns true if Cache is in a state where it is up but not necessarily cleanly
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

// Returns true when the instance is down
func (iis InstanceStatus) Down() bool {
	switch iis {
	default:
		return false
	case
		InstanceStatusDown:
		return true
	}
}

// Returns true when a bypass is required to stop the instance
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
