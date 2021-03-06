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

	"github.com/sirupsen/logrus"
)

// EmptyResourceClient is a ResourceClient with always exists and dose nothing
var EmptyResourceClient = emptyResourceClient{}

type resourceClient struct {
	id     string
	name   string
	logger *logrus.Entry
}

func (client *resourceClient) ID() (string, error) {
	return client.id, nil
}
func (client *resourceClient) Name() (string, error) {
	return client.name, nil
}

func (client *resourceClient) Exists() (bool, error) {
	return false, fmt.Errorf("Exists not supported")
}
func (client *resourceClient) Create(dryRun bool) (changed bool, err error) {
	return changed, fmt.Errorf("Create not supported")
}
func (client *resourceClient) Upgrade(dryRun bool) (changed bool, err error) {
	return changed, fmt.Errorf("Upgrade not supported")
}
func (client *resourceClient) Delete(dryRun bool) (changed bool, err error) {
	return changed, fmt.Errorf("Delete not supported")
}

type namespacedResourceClient struct {
	resourceClient
	namespaceID string
	namespace   string
	project     ProjectClient
}

func (client *namespacedResourceClient) NamespaceID() (string, error) {
	if client.namespace == "" {
		return "", nil
	}

	if client.namespaceID != "" {
		return client.namespaceID, nil
	}
	var namespace NamespaceClient
	var err error
	if namespace, err = client.project.Namespace(client.namespace); err != nil {
		client.logger.WithError(err).Error("Failed to read namespaceID")
		return "", fmt.Errorf("Failed to read namespaceID, %v", err)
	}
	if client.namespaceID, err = namespace.ID(); err != nil {
		client.logger.WithError(err).Error("Failed to read namespaceID")
		return "", fmt.Errorf("Failed to read namespaceID, %v", err)
	}
	return client.namespaceID, nil
}
func (client *namespacedResourceClient) Namespace() (string, error) {
	return client.namespace, nil
}

type emptyResourceClient struct{}

func (client emptyResourceClient) ID() (string, error) {
	return "", nil
}
func (client emptyResourceClient) Type() string {
	return "EmptyResource"
}
func (client emptyResourceClient) Name() (string, error) {
	return "", nil
}

func (client emptyResourceClient) Exists() (bool, error) {
	return true, nil
}
func (client emptyResourceClient) Create(dryRun bool) (changed bool, err error) {
	return
}
func (client emptyResourceClient) Upgrade(dryRun bool) (changed bool, err error) {
	return
}
func (client emptyResourceClient) Delete(dryRun bool) (changed bool, err error) {
	return
}
