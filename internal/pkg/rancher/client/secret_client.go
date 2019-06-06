// Copyright © 2019 Bitgrip <berlin@bitgrip.de>
//
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

package client

import (
	"fmt"
	"reflect"

	projectModel "github.com/bitgrip/cattlectl/internal/pkg/rancher/project/model"
	"github.com/rancher/norman/types"
	backendProjectClient "github.com/rancher/types/client/project/v3"
	"github.com/sirupsen/logrus"
)

func newSecretpClientWithData(
	secret projectModel.ConfigMap,
	namespace string,
	project ProjectClient,
	logger *logrus.Entry,
) (ConfigMapClient, error) {
	result, err := newSecretClient(
		secret.Name,
		namespace,
		project,
		logger,
	)
	if err != nil {
		return nil, err
	}
	err = result.SetData(secret)
	return result, err
}

func newSecretClient(
	name, namespace string,
	project ProjectClient,
	logger *logrus.Entry,
) (ConfigMapClient, error) {
	return &secretClient{
		namespacedResourceClient: namespacedResourceClient{
			resourceClient: resourceClient{
				name:   name,
				logger: logger.WithField("secret_name", name).WithField("namespace", namespace),
			},
			namespace: namespace,
			project:   project,
		},
	}, nil
}

type secretClient struct {
	namespacedResourceClient
	secret projectModel.ConfigMap
}

func (client *secretClient) Exists() (bool, error) {
	if client.namespace != "" {
		return client.existsInNamespace()
	}
	return client.existsInProject()
}

func (client *secretClient) existsInProject() (bool, error) {
	backendClient, err := client.project.backendProjectClient()
	if err != nil {
		return false, err
	}
	collection, err := backendClient.Secret.List(&types.ListOpts{
		Filters: map[string]interface{}{
			"name": client.name,
		},
	})
	if nil != err {
		client.logger.WithError(err).Error("Failed to read secret list")
		return false, fmt.Errorf("Failed to read secret list, %v", err)
	}
	for _, item := range collection.Data {
		if item.Name == client.name && item.NamespaceId == client.namespaceID {
			return true, nil
		}
	}
	client.logger.Debug("Secret not found")
	return false, nil
}

func (client *secretClient) existsInNamespace() (bool, error) {
	backendClient, err := client.project.backendProjectClient()
	if err != nil {
		return false, err
	}
	namespaceID, err := client.NamespaceID()
	if err != nil {
		return false, err
	}
	collection, err := backendClient.NamespacedSecret.List(&types.ListOpts{
		Filters: map[string]interface{}{
			"name":        client.name,
			"namespaceId": namespaceID,
		},
	})
	if nil != err {
		client.logger.WithError(err).Error("Failed to read secret list")
		return false, fmt.Errorf("Failed to read secret list, %v", err)
	}
	for _, item := range collection.Data {
		if item.Name == client.name && item.NamespaceId == namespaceID {
			return true, nil
		}
	}
	client.logger.Debug("Secret not found")
	return false, nil
}

func (client *secretClient) Create() error {
	if client.namespace != "" {
		return client.createInNamespace()
	}
	return client.createInProject()
}

func (client *secretClient) createInProject() error {
	backendClient, err := client.project.backendProjectClient()
	if err != nil {
		return err
	}
	projectID, err := client.project.ID()
	if err != nil {
		return err
	}
	client.logger.Info("Create new Secret")
	labels := make(map[string]string)
	labels["cattlectl.io/hash"] = hashOf(client.secret)
	newSecret := &backendProjectClient.Secret{
		Name:      client.secret.Name,
		Labels:    labels,
		Data:      client.secret.Data,
		ProjectID: projectID,
	}

	_, err = backendClient.Secret.Create(newSecret)
	return err
}

func (client *secretClient) createInNamespace() error {
	backendClient, err := client.project.backendProjectClient()
	if err != nil {
		return err
	}
	namespaceID, err := client.NamespaceID()
	if err != nil {
		return err
	}
	projectID, err := client.project.ID()
	if err != nil {
		return fmt.Errorf("Failed to read namespace ID, %v", err)
	}
	client.logger.Info("Create new Secret")
	labels := make(map[string]string)
	labels["cattlectl.io/hash"] = hashOf(client.secret)
	newSecret := &backendProjectClient.NamespacedSecret{
		Name:        client.secret.Name,
		Labels:      labels,
		Data:        client.secret.Data,
		NamespaceId: namespaceID,
		ProjectID:   projectID,
	}

	_, err = backendClient.NamespacedSecret.Create(newSecret)
	return err
}

func (client *secretClient) Upgrade() error {
	if client.namespace != "" {
		return client.upgradeInNamespace()
	}
	return client.upgradeInProject()
}

func (client *secretClient) upgradeInProject() error {
	backendClient, err := client.project.backendProjectClient()
	if err != nil {
		return err
	}
	collection, err := backendClient.Secret.List(&types.ListOpts{
		Filters: map[string]interface{}{
			"name": client.name,
		},
	})
	if nil != err {
		client.logger.WithError(err).Error("Failed to read secret list")
		return fmt.Errorf("Failed to read secret list, %v", err)
	}

	if len(collection.Data) == 0 {
		return fmt.Errorf("Secret %v not found", client.name)
	}
	existingSecret := collection.Data[0]
	if isProjectSecretUnchanged(existingSecret, client.secret) {
		client.logger.Debug("Skip upgrade secret - no changes")
		return nil
	}
	client.logger.Info("Upgrade Secret")
	existingSecret.Data = client.secret.Data

	_, err = backendClient.Secret.Replace(&existingSecret)
	return err
}

func (client *secretClient) upgradeInNamespace() error {
	backendClient, err := client.project.backendProjectClient()
	if err != nil {
		return err
	}
	namespaceID, err := client.NamespaceID()
	if err != nil {
		return err
	}
	collection, err := backendClient.NamespacedSecret.List(&types.ListOpts{
		Filters: map[string]interface{}{
			"name":        client.name,
			"namespaceId": namespaceID,
		},
	})
	if nil != err {
		client.logger.WithError(err).Error("Failed to read secret list")
		return fmt.Errorf("Failed to read secret list, %v", err)
	}

	if len(collection.Data) == 0 {
		return fmt.Errorf("Secret %v not found", client.name)
	}
	existingSecret := collection.Data[0]
	if isNamespacedSecretUnchanged(existingSecret, client.secret) {
		client.logger.Debug("Skip upgrade secret - no changes")
		return nil
	}
	client.logger.Info("Upgrade Secret")
	existingSecret.Data = client.secret.Data

	_, err = backendClient.NamespacedSecret.Replace(&existingSecret)
	return err
}

func (client *secretClient) Data() (projectModel.ConfigMap, error) {
	return client.secret, nil
}

func (client *secretClient) SetData(secret projectModel.ConfigMap) error {
	client.name = secret.Name
	client.secret = secret
	return nil
}

func isProjectSecretUnchanged(existingSecret backendProjectClient.Secret, secret projectModel.ConfigMap) bool {
	return reflect.DeepEqual(existingSecret.Data, secret.Data)
}

func isNamespacedSecretUnchanged(existingSecret backendProjectClient.NamespacedSecret, secret projectModel.ConfigMap) bool {
	return reflect.DeepEqual(existingSecret.Data, secret.Data)
}