// Copyright 2019 The Terraformer Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package azure

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/go-autorest/autorest"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils"
	"github.com/GoogleCloudPlatform/terraformer/terraformutils/providerwrapper"
	"github.com/hashicorp/go-azure-helpers/authentication"
	"github.com/hashicorp/go-azure-helpers/sender"
)

type AzureProvider struct { //nolint
	terraformutils.Provider
	config        authentication.Config
	authorizer    autorest.Authorizer
	resourceGroup string
}

func (p *AzureProvider) setEnvConfig() error {
	subscriptionID := os.Getenv("ARM_SUBSCRIPTION_ID")
	if subscriptionID == "" {
		return errors.New("set ARM_SUBSCRIPTION_ID env var")
	}
	builder := &authentication.Builder{
		ClientID:                 os.Getenv("ARM_CLIENT_ID"),
		SubscriptionID:           subscriptionID,
		TenantID:                 os.Getenv("ARM_TENANT_ID"),
		Environment:              os.Getenv("ARM_ENVIRONMENT"),
		ClientSecret:             os.Getenv("ARM_CLIENT_SECRET"),
		SupportsAzureCliToken:    true,
		SupportsClientSecretAuth: true,
		SupportsClientCertAuth:   true,
		ClientCertPath:           os.Getenv("ARM_CLIENT_CERTIFICATE_PATH"),
		ClientCertPassword:       os.Getenv("ARM_CLIENT_CERTIFICATE_PASSWORD"),

		/*
		   // Managed Service Identity Auth
		   SupportsManagedServiceIdentity bool
		   MsiEndpoint                    string
		*/
	}

	if builder.Environment == "" {
		builder.Environment = "public"
	}
	config, err := builder.Build()
	if err != nil {
		return nil
	}
	p.config = *config

	return nil
}

func (p *AzureProvider) getAuthorizer() (autorest.Authorizer, error) {
	env, err := authentication.DetermineEnvironment(p.config.Environment)
	if err != nil {
		return nil, err
	}

	oauthConfig, err := p.config.BuildOAuthConfig(env.ActiveDirectoryEndpoint)
	if err != nil {
		return nil, err
	}

	if oauthConfig == nil {
		return nil, fmt.Errorf("Unable to configure OAuthConfig for tenant %s", p.config.TenantID)
	}

	sender := sender.BuildSender("AzureRM")

	auth, err := p.config.GetAuthorizationToken(sender, oauthConfig, env.ResourceManagerEndpoint)
	if err != nil {
		return nil, err
	}

	return auth, nil
}

func (p *AzureProvider) Init(args []string) error {
	err := p.setEnvConfig()
	if err != nil {
		return err
	}

	authorizer, err := p.getAuthorizer()
	if err != nil {
		return err
	}
	p.authorizer = authorizer
	p.resourceGroup = args[0]

	return nil
}

func (p *AzureProvider) GetName() string {
	return "azurerm"
}

func (p *AzureProvider) GetProviderData(arg ...string) map[string]interface{} {
	version := providerwrapper.GetProviderVersion(p.GetName())
	if strings.Contains(version, "v2.") {
		return map[string]interface{}{
			"provider": map[string]interface{}{
				"azurerm": map[string]interface{}{
					// NOTE:
					// Workaround for azurerm v2 provider changes
					// Tested with azurerm_resource_group under v2.17.0
					// https://github.com/terraform-providers/terraform-provider-azurerm/issues/5866#issuecomment-594239342
					// https://github.com/hashicorp/terraform/issues/24200#issuecomment-594745861
					"features": map[string]interface{}{},
				},
			},
		}
	}
	return map[string]interface{}{
		"provider": map[string]interface{}{
			"azurerm": map[string]interface{}{
				"version": version,
			},
		},
	}
}

func (AzureProvider) GetResourceConnections() map[string]map[string][]string {
	return map[string]map[string][]string{
		"analysis": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"app_service": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"cosmosdb": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"container": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"database": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"databricks": {
			"resource_group": []string{
				"resource_group_name", "name",
				"managed_resource_group_name", "name",
			},
			"storage_account": []string{"storage_account_name", "name"},
			"subnet": []string{
				"public_subnet_name", "name",
				"private_subnet_name", "name",
			},
			"virtual_network": []string{"virtual_network_id", "id"},
		},
		"data_factory": {
			"resource_group": []string{"resource_group_name", "name"},
			"data_factory":   []string{"data_factory_name", "name"},
			"keyvault":       []string{"keyvault_id", "id"},
		},
		"disk": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"dns": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"eventhub": {
			"resource_group": []string{"resource_group_name", "name"},
			"eventhub": []string{
				"eventhub_name", "name",
				"namespace_name", "name",
			},
		},
		"keyvault": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"load_balancer": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"network_interface": {
			"resource_group": []string{"resource_group_name", "name"},
			"subnet":         []string{"subnet_id", "id"},
		},
		"network_security_group": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"private_dns": {
			"resource_group":  []string{"resource_group_name", "name"},
			"virtual_network": []string{"virtual_network_id", "id"},
		},
		"private_endpoint": {
			"resource_group": []string{"resource_group_name", "name"},
			"subnet":         []string{"subnet_id", "id"},
		},
		"public_ip": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"purview": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"redis": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"scaleset": {
			"resource_group": []string{"resource_group_name", "name"},
		},
		"storage_account": {
			"resource_group":  []string{"resource_group_name", "name"},
			"virtual_network": []string{"virtual_network_subnet_ids", "id"},
		},
		"storage_blob": {
			"storage_account":   []string{"storage_account_name", "name"},
			"storage_container": []string{"storage_container_name", "name"},
		},
		"storage_container": {
			"storage_account": []string{"storage_account_name", "name"},
		},
		"synapse": {
			"resource_group": []string{
				"resource_group_name", "name",
				"managed_resource_group_name", "name",
			},
			"synapse": []string{"synapse_workspace_id", "id"},
		},
		"subnet": {
			"resource_group":         []string{"resource_group_name", "name"},
			"virtual_network":        []string{"virtual_network_name", "name"},
			"network_security_group": []string{"network_security_group_id", "id"},
			"subnet":                 []string{"subnet_id", "id"},
		},
		"virtual_machine": {
			"resource_group":    []string{"resource_group_name", "name"},
			"network_interface": []string{"network_interface_ids", "id"},
		},
		"virtual_network": {
			"resource_group": []string{"resource_group_name", "name"},
		},
	}
}

func (p *AzureProvider) GetSupportedService() map[string]terraformutils.ServiceGenerator {
	return map[string]terraformutils.ServiceGenerator{
		"analysis":                             &AnalysisGenerator{},
		"app_service":                          &AppServiceGenerator{},
		"cosmosdb":                             &CosmosDBGenerator{},
		"container":                            &ContainerGenerator{},
		"database":                             &DatabasesGenerator{},
		"databricks":                           &DatabricksGenerator{},
		"data_factory":                         &DataFactoryGenerator{},
		"disk":                                 &DiskGenerator{},
		"dns":                                  &DNSGenerator{},
		"eventhub":                             &EventHubGenerator{},
		"keyvault":                             &KeyVaultGenerator{},
		"load_balancer":                        &LoadBalancerGenerator{},
		"network_interface":                    &NetworkInterfaceGenerator{},
		"network_security_group":               &NetworkSecurityGroupGenerator{},
		"private_dns":                          &PrivateDNSGenerator{},
		"private_endpoint":                     &PrivateEndpointGenerator{},
		"public_ip":                            &PublicIPGenerator{},
		"purview":                              &PurviewGenerator{},
		"redis":                                &RedisGenerator{},
		"resource_group":                       &ResourceGroupGenerator{},
		"scaleset":                             &ScaleSetGenerator{},
		"security_center_contact":              &SecurityCenterContactGenerator{},
		"security_center_subscription_pricing": &SecurityCenterSubscriptionPricingGenerator{},
		"storage_account":                      &StorageAccountGenerator{},
		"storage_blob":                         &StorageBlobGenerator{},
		"storage_container":                    &StorageContainerGenerator{},
		"synapse":                              &SynapseGenerator{},
		"subnet":                               &SubnetGenerator{},
		"virtual_machine":                      &VirtualMachineGenerator{},
		"virtual_network":                      &VirtualNetworkGenerator{},
	}
}

func (p *AzureProvider) InitService(serviceName string, verbose bool) error {
	var isSupported bool
	if _, isSupported = p.GetSupportedService()[serviceName]; !isSupported {
		return errors.New("azurerm: " + serviceName + " not supported service")
	}
	p.Service = p.GetSupportedService()[serviceName]
	p.Service.SetName(serviceName)
	p.Service.SetVerbose(verbose)
	p.Service.SetProviderName(p.GetName())
	p.Service.SetArgs(map[string]interface{}{
		"config":         p.config,
		"authorizer":     p.authorizer,
		"resource_group": p.resourceGroup,
	})
	return nil
}

func (p *AzureService) getClientArgs() (subscriptionID string, resourceGroup string, authorizer autorest.Authorizer) {
	subs := p.Args["config"].(authentication.Config).SubscriptionID
	auth := p.Args["authorizer"].(autorest.Authorizer)
	resg := p.Args["resource_group"].(string)
	return subs, resg, auth
}

func (p *AzureService) AppendSimpleResource(id string, resourceName string, resourceType string) {
	newResource := terraformutils.NewSimpleResource(id, resourceName, resourceType, p.ProviderName, []string{})
	p.Resources = append(p.Resources, newResource)
}

func (p *AzureService) appendSimpleAssociation(id string, linkedResourceName string, resourceName *string, resourceType string, attributes map[string]string) {
	var resourceName2 string
	if resourceName != nil {
		resourceName2 = *resourceName
	} else {
		resourceName0 := strings.ReplaceAll(resourceType, "azurerm_", "")
		resourceName1 := resourceName0[strings.IndexByte(resourceName0, '_'):]
		resourceName2 = linkedResourceName + resourceName1
	}
	newResource := terraformutils.NewResource(
		id, resourceName2, resourceType, p.ProviderName, attributes,
		[]string{"name"},
		map[string]interface{}{},
	)
	p.Resources = append(p.Resources, newResource)
}
