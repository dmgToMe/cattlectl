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

func newCronJobClientWithData(
	cronJob projectModel.CronJob,
	namespace string,
	project ProjectClient,
	logger *logrus.Entry,
) (CronJobClient, error) {
	result, err := newCronJobClient(
		cronJob.Name,
		namespace,
		project,
		logger,
	)
	if err != nil {
		return nil, err
	}
	err = result.SetData(cronJob)
	return result, err
}

func newCronJobClient(
	name, namespace string,
	project ProjectClient,
	logger *logrus.Entry,
) (CronJobClient, error) {
	return &cronJobClient{
		namespacedResourceClient: namespacedResourceClient{
			resourceClient: resourceClient{
				name:   name,
				logger: logger.WithField("cronJob_name", name).WithField("namespace", namespace),
			},
			namespace: namespace,
			project:   project,
		},
	}, nil
}

type cronJobClient struct {
	namespacedResourceClient
	cronJob projectModel.CronJob
}

func (client *cronJobClient) Type() string {
	return rancherModel.CronJobKind
}

func (client *cronJobClient) Exists() (bool, error) {
	backendClient, err := client.project.backendProjectClient()
	if err != nil {
		return false, err
	}
	namespaceID, err := client.NamespaceID()
	if err != nil {
		return false, err
	}
	collection, err := backendClient.CronJob.List(&types.ListOpts{
		Filters: map[string]interface{}{
			"name":        client.name,
			"namespaceId": namespaceID,
		},
	})
	if nil != err {
		client.logger.WithError(err).Error("Failed to read cronJob list")
		return false, fmt.Errorf("Failed to read cronJob list, %v", err)
	}
	for _, item := range collection.Data {
		if item.Name == client.name && item.NamespaceId == namespaceID {
			return true, nil
		}
	}
	client.logger.Debug("CronJob not found")
	return false, nil
}

func (client *cronJobClient) Create(dryRun bool) (changed bool, err error) {
	backendClient, err := client.project.backendProjectClient()
	if err != nil {
		return
	}
	namespaceID, err := client.NamespaceID()
	if err != nil {
		return
	}
	client.logger.Info("Create new cronJob")
	pattern, err := projectModel.ConvertCronJobToProjectAPI(client.cronJob)
	if err != nil {
		return
	}
	pattern.NamespaceId = namespaceID

	if dryRun {
		client.logger.WithField("object", pattern).Info("Do Dry-Run Create")
	} else {
		_, err = backendClient.CronJob.Create(&pattern)
	}
	return err == nil, err
}

func (client *cronJobClient) Upgrade(dryRun bool) (changed bool, err error) {
	client.logger.Warn("Skip change existing cronjob")
	return
}

func (client *cronJobClient) Data() (projectModel.CronJob, error) {
	return client.cronJob, nil
}

func (client *cronJobClient) SetData(cronJob projectModel.CronJob) error {
	client.name = cronJob.Name
	client.cronJob = cronJob
	return nil
}
