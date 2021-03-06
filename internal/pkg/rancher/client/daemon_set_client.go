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

	projectModel "github.com/bitgrip/cattlectl/internal/pkg/rancher/cluster/project/model"
	rancherModel "github.com/bitgrip/cattlectl/internal/pkg/rancher/model"
	"github.com/rancher/norman/types"
	"github.com/sirupsen/logrus"
)

func newDaemonSetClientWithData(
	daemonSet projectModel.DaemonSet,
	namespace string,
	project ProjectClient,
	logger *logrus.Entry,
) (DaemonSetClient, error) {
	result, err := newDaemonSetClient(
		daemonSet.Name,
		namespace,
		project,
		logger,
	)
	if err != nil {
		return nil, err
	}
	err = result.SetData(daemonSet)
	return result, err
}

func newDaemonSetClient(
	name, namespace string,
	project ProjectClient,
	logger *logrus.Entry,
) (DaemonSetClient, error) {
	return &daemonSetClient{
		namespacedResourceClient: namespacedResourceClient{
			resourceClient: resourceClient{
				name:   name,
				logger: logger.WithField("daemonSet_name", name).WithField("namespace", namespace),
			},
			namespace: namespace,
			project:   project,
		},
	}, nil
}

type daemonSetClient struct {
	namespacedResourceClient
	daemonSet projectModel.DaemonSet
}

func (client *daemonSetClient) Type() string {
	return rancherModel.DaemonSetKind
}

func (client *daemonSetClient) Exists() (bool, error) {
	backendClient, err := client.project.backendProjectClient()
	if err != nil {
		return false, err
	}
	namespaceID, err := client.NamespaceID()
	if err != nil {
		return false, err
	}
	collection, err := backendClient.DaemonSet.List(&types.ListOpts{
		Filters: map[string]interface{}{
			"name":        client.name,
			"namespaceId": namespaceID,
		},
	})
	if nil != err {
		client.logger.WithError(err).Error("Failed to read daemonSet list")
		return false, fmt.Errorf("Failed to read daemonSet list, %v", err)
	}
	for _, item := range collection.Data {
		if item.Name == client.name && item.NamespaceId == namespaceID {
			return true, nil
		}
	}
	client.logger.Debug("DaemonSet not found")
	return false, nil
}

func (client *daemonSetClient) Create(dryRun bool) (changed bool, err error) {
	backendClient, err := client.project.backendProjectClient()
	if err != nil {
		return
	}
	namespaceID, err := client.NamespaceID()
	if err != nil {
		return
	}
	client.logger.Info("Create new daemonSet")
	pattern, err := projectModel.ConvertDaemonSetToProjectAPI(client.daemonSet)
	if err != nil {
		return
	}
	pattern.NamespaceId = namespaceID

	if dryRun {
		client.logger.WithField("object", pattern).Info("Do Dry-Run Create")
	} else {
		_, err = backendClient.DaemonSet.Create(&pattern)
	}
	return err == nil, err
}

func (client *daemonSetClient) Upgrade(dryRun bool) (changed bool, err error) {
	client.logger.Warn("Skip change existing daemonset")
	return
}

func (client *daemonSetClient) Data() (projectModel.DaemonSet, error) {
	return client.daemonSet, nil
}

func (client *daemonSetClient) SetData(daemonSet projectModel.DaemonSet) error {
	client.name = daemonSet.Name
	client.daemonSet = daemonSet
	return nil
}
