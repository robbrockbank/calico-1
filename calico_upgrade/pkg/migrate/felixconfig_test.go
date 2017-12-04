// Copyright (c) 2017 Tigera, Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package migrate_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	apiv3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
	"github.com/projectcalico/libcalico-go/lib/backend/model"
	"github.com/projectcalico/libcalico-go/lib/backend/syncersv1/updateprocessors"
	"github.com/projectcalico/calico/calico_upgrade/pkg/migrate"
	"github.com/sirupsen/logrus"
)

var _ = Describe("Test felix configuration upgrade", func() {
	// Define some common values
	perNodeFelixKey := model.ResourceKey{
		Kind: apiv3.KindFelixConfiguration,
		Name: "node.mynode",
	}
	globalFelixKey := model.ResourceKey{
		Kind: apiv3.KindFelixConfiguration,
		Name: "default",
	}
	globalClusterKey := model.ResourceKey{
		Kind: apiv3.KindClusterInformation,
		Name: "default",
	}

	It("should handle different field types being assigned", func() {
		cc := updateprocessors.NewFelixConfigUpdateProcessor()
		By("converting a per-node felix KVPair with certain values and checking for the correct number of fields")
		res := apiv3.NewFelixConfiguration()
		int1 := int(12345)
		bool1 := false
		uint1 := uint32(1313)
		res.Spec.RouteRefreshIntervalSecs = &int1
		res.Spec.InterfacePrefix = "califoobar"
		res.Spec.IPIPEnabled = &bool1
		res.Spec.IptablesMarkMask = &uint1
		res.Spec.FailsafeInboundHostPorts = &[]apiv3.ProtoPort{}
		res.Spec.FailsafeOutboundHostPorts = &[]apiv3.ProtoPort{
			{
				Protocol: "TCP",
				Port:     1234,
			},
			{
				Protocol: "UDP",
				Port:     22,
			},
			{
				Protocol: "TCP",
				Port:     65535,
			},
		}

		// Convert v3 felix configuration into per host kvps.
		kvps, err := cc.Process(&model.KVPair{
			Key:   perNodeFelixKey,
			Value: res,
		})
		Expect(err).NotTo(HaveOccurred())

		// Convert back
		fc := &migrate.FelixConfig{}
		hostConfig := apiv3.NewFelixConfiguration()
		hostConfig.Name = fmt.Sprint("node.%s", "mynode")
		_, _, err := fc.convertFelixConfigV1KVToV3Resource(kvps, hostConfig)
		Expect(err).NotTo(HaveOccurred())

		Expect(res).To(Equal(hostConfig)) // DeepEqual

		// Convert v3 felix configuration into global kvps.
		kvps, err = cc.Process(&model.KVPair{
			Key:   globalFelixKey,
			Value: res,
		})
		Expect(err).NotTo(HaveOccurred())

		// Convert back
		globalConfig := apiv3.NewFelixConfiguration()
		globalConfig.Name = "default"
		_, _, err := fc.convertFelixConfigV1KVToV3Resource(kvps, globalConfig)
		Expect(err).NotTo(HaveOccurred())

		Expect(res).To(Equal(globalConfig)) // DeepEqual
	})

	It("should handle cluster config string slice field", func() {
		cc := updateprocessors.NewClusterInfoUpdateProcessor()
		By("converting a global cluster info KVPair with values assigned")
		res := apiv3.NewClusterInformation()
		res.Spec.ClusterGUID = "abcedfg"
		res.Spec.ClusterType = "Mesos,K8s"
		ready := true
		res.Spec.DatastoreReady = &ready

		kvps, err := cc.Process(&model.KVPair{
			Key:   globalClusterKey,
			Value: res,
		})
		Expect(err).NotTo(HaveOccurred())

		fc := &migrate.FelixConfig{}
		clusterInfo := apiv3.NewClusterInformation()
		clusterInfo.Name = "default"
		_, _, err = fc.convertFelixConfigV1KVToV3Resource(kvps, clusterInfo); err != nil {
		Expect(err).NotTo(HaveOccurred())

		Expect(res).To(Equal(clusterInfo)) // DeepEqual

	})
})