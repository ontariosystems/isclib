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

package isclib_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/ontariosystems/isclib"
)

var _ = Describe("InstanceStatus", func() {
	DescribeTable("Status value", func(input string, handled, running, up, down, bypass bool) {
		status := isclib.InstanceStatus(input)
		Expect(status.Handled()).To(Equal(handled), "Handled")
		Expect(status.Ready()).To(Equal(running), "Ready")
		Expect(status.Up()).To(Equal(up), "Up")
		Expect(status.Down()).To(Equal(down), "Down")
		Expect(status.RequiresBypass()).To(Equal(bypass), "Bypass")
	},
		Entry("is blank", "", false, false, false, false, false),
		Entry("is unknown", "unknown", false, false, false, false, false),
		Entry("is running", "running", true, true, true, false, false),
		Entry("is sign-on inhibited", "sign-on inhibited", true, false, true, false, true),
		Entry("is in primary transition", "sign-on inhibited:primary transition", true, false, true, false, true),
		Entry("is down", "down", true, false, false, true, false),
		Entry("is running without cache.ids", "running on node ? (cache.ids missing)", true, true, true, false, false),
	)
})
