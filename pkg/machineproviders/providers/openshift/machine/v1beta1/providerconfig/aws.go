/*
Copyright 2022 Red Hat, Inc.

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

package providerconfig

import (
	"encoding/json"
	"fmt"

	configv1 "github.com/openshift/api/config/v1"
	machinev1 "github.com/openshift/api/machine/v1"
	machinev1beta1 "github.com/openshift/api/machine/v1beta1"
	"github.com/openshift/cluster-control-plane-machine-set-operator/pkg/machineproviders/providers/openshift/machine/v1beta1/failuredomain"
	"k8s.io/apimachinery/pkg/runtime"
)

// AWSProviderConfig holds the provider spec of an AWS Machine.
// It allows external code to extract and inject failure domain information,
// as well as gathering the stored config.
type AWSProviderConfig struct {
	providerConfig machinev1beta1.AWSMachineProviderConfig
}

// InjectFailureDomain returns a new AWSProviderConfig configured with the failure domain
// information provided.
func (a AWSProviderConfig) InjectFailureDomain(fd machinev1.AWSFailureDomain) AWSProviderConfig {
	newAWSProviderConfig := a

	newAWSProviderConfig.providerConfig.Placement.AvailabilityZone = fd.Placement.AvailabilityZone
	newAWSProviderConfig.providerConfig.Subnet = failuredomain.ConvertAWSResourceReferenceToBeta1(fd.Subnet)

	return newAWSProviderConfig
}

// ExtractFailureDomain returns an AWSFailureDomain based on the failure domain
// information stored within the AWSProviderConfig.
func (a AWSProviderConfig) ExtractFailureDomain() machinev1.AWSFailureDomain {
	return machinev1.AWSFailureDomain{
		Placement: machinev1.AWSFailureDomainPlacement{
			AvailabilityZone: a.providerConfig.Placement.AvailabilityZone,
		},
		Subnet: failuredomain.ConvertAWSResourceReferenceToV1(a.providerConfig.Subnet),
	}
}

// Config returns the stored AWSMachineProviderConfig.
func (a AWSProviderConfig) Config() machinev1beta1.AWSMachineProviderConfig {
	return a.providerConfig
}

// newAWSProviderConfig creates an AWS type ProviderConfig from the raw extension.
// It should return an error if the provided RawExtension does not represent
// an AWSMachineProviderConfig.
func newAWSProviderConfig(raw *runtime.RawExtension) (ProviderConfig, error) {
	awsMachineProviderConfig := machinev1beta1.AWSMachineProviderConfig{}
	if err := json.Unmarshal(raw.Raw, &awsMachineProviderConfig); err != nil {
		return nil, fmt.Errorf("could not unmarshal provider spec: %w", err)
	}

	awsProviderConfig := AWSProviderConfig{
		providerConfig: awsMachineProviderConfig,
	}

	config := providerConfig{
		platformType: configv1.AWSPlatformType,
		aws:          awsProviderConfig,
	}

	return config, nil
}
