//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ctl

import (
	"fmt"

	"github.com/bitgrip/cattlectl/internal/pkg/config"
	"github.com/bitgrip/cattlectl/internal/pkg/rancher/client"
	"github.com/sirupsen/logrus"
)

var (
	deletableProjectResouceTypes = map[string]func(string, string, string, config.Config) (bool, error){
		"namespace":         deleteNamespace,
		"certificate":       deleteCertificate,
		"config-map":        deleteConfigMap,
		"docker-credential": deleteDockerCredential,
		"secret":            deleteSecret,
		"app":               deleteApp,
		"job":               deleteJob,
		"cron-job":          deleteCronJob,
		"deployment":        deleteDeployment,
		"daemon-set":        deleteDaemonSet,
		"stateful-set":      deleteStatefulSet,
	}
)

// DeleteProjectResouce is deleting one project resource from project
//
// * projectName: the project to delete the resource from
// * resourceType: the type of the resource to delete
// * name: the name of the resource to delete
func DeleteProjectResouce(projectName, namespace, kind, name string, config config.Config) (bool, error) {
	deleteFunc, supportedType := deletableProjectResouceTypes[kind]
	if !supportedType {
		return false, fmt.Errorf("Not supported resouce type [%s]", kind)
	}
	return deleteFunc(projectName, namespace, name, config)
}

func deleteNamespace(projectName, _namespace, name string, config config.Config) (deleted bool, err error) {
	_, _, projectClient, err := getProjectClient(projectName, config)
	if err != nil {
		return
	}

	namespace, err := projectClient.Namespace(name)
	if err != nil {
		return
	}

	return deleteProjectResouce(namespace, config.ClusterName(), projectName, "namespace", name, config.DryRun())
}

func deleteCertificate(projectName, namespace, name string, config config.Config) (deleted bool, err error) {
	_, _, projectClient, err := getProjectClient(projectName, config)
	if err != nil {
		return
	}

	certificate, err := projectClient.Certificate(name, namespace)
	if err != nil {
		return
	}

	if namespace == "" {
		return deleteProjectResouce(certificate, config.ClusterName(), projectName, "certificate", name, config.DryRun())
	}
	return deleteNamespaceResouce(certificate, config.ClusterName(), projectName, namespace, "certificate", name, config.DryRun())
}

func deleteConfigMap(projectName, namespace, name string, config config.Config) (deleted bool, err error) {
	_, _, projectClient, err := getProjectClient(projectName, config)
	if err != nil {
		return
	}

	configMap, err := projectClient.ConfigMap(name, namespace)
	if err != nil {
		return
	}

	return deleteNamespaceResouce(configMap, config.ClusterName(), projectName, namespace, "config-map", name, config.DryRun())
}

func deleteDockerCredential(projectName, namespace, name string, config config.Config) (deleted bool, err error) {
	_, _, projectClient, err := getProjectClient(projectName, config)
	if err != nil {
		return
	}

	dockerCredential, err := projectClient.DockerCredential(name, namespace)
	if err != nil {
		return
	}

	if namespace == "" {
		return deleteProjectResouce(dockerCredential, config.ClusterName(), projectName, "docker-credential", name, config.DryRun())
	}
	return deleteNamespaceResouce(dockerCredential, config.ClusterName(), projectName, namespace, "docker-credential", name, config.DryRun())
}

func deleteSecret(projectName, namespace, name string, config config.Config) (deleted bool, err error) {
	_, _, projectClient, err := getProjectClient(projectName, config)
	if err != nil {
		return
	}

	secret, err := projectClient.Secret(name, namespace)
	if err != nil {
		return
	}

	if namespace == "" {
		return deleteProjectResouce(secret, config.ClusterName(), projectName, "secret", name, config.DryRun())
	}
	return deleteNamespaceResouce(secret, config.ClusterName(), projectName, namespace, "secret", name, config.DryRun())
}

func deleteApp(projectName, namespace, name string, config config.Config) (deleted bool, err error) {
	_, _, projectClient, err := getProjectClient(projectName, config)
	if err != nil {
		return
	}

	app, err := projectClient.App(name)
	if err != nil {
		return
	}

	return deleteProjectResouce(app, config.ClusterName(), projectName, "app", name, config.DryRun())
}

func deleteJob(projectName, namespace, name string, config config.Config) (deleted bool, err error) {
	_, _, projectClient, err := getProjectClient(projectName, config)
	if err != nil {
		return
	}

	job, err := projectClient.Job(name, namespace)
	if err != nil {
		return
	}

	return deleteNamespaceResouce(job, config.ClusterName(), projectName, namespace, "job", name, config.DryRun())
}

func deleteCronJob(projectName, namespace, name string, config config.Config) (deleted bool, err error) {
	_, _, projectClient, err := getProjectClient(projectName, config)
	if err != nil {
		return
	}

	cronJob, err := projectClient.CronJob(name, namespace)
	if err != nil {
		return
	}

	return deleteNamespaceResouce(cronJob, config.ClusterName(), projectName, namespace, "cron-job", name, config.DryRun())
}

func deleteDeployment(projectName, namespace, name string, config config.Config) (deleted bool, err error) {
	_, _, projectClient, err := getProjectClient(projectName, config)
	if err != nil {
		return
	}

	deployment, err := projectClient.Deployment(name, namespace)
	if err != nil {
		return
	}

	return deleteNamespaceResouce(deployment, config.ClusterName(), projectName, namespace, "deployment", name, config.DryRun())
}

func deleteDaemonSet(projectName, namespace, name string, config config.Config) (deleted bool, err error) {
	_, _, projectClient, err := getProjectClient(projectName, config)
	if err != nil {
		return
	}

	daemonSet, err := projectClient.DaemonSet(name, namespace)
	if err != nil {
		return
	}

	return deleteNamespaceResouce(daemonSet, config.ClusterName(), projectName, namespace, "daemon-set", name, config.DryRun())
}

func deleteStatefulSet(projectName, namespace, name string, config config.Config) (deleted bool, err error) {
	_, _, projectClient, err := getProjectClient(projectName, config)
	if err != nil {
		return
	}

	statefulSet, err := projectClient.StatefulSet(name, namespace)
	if err != nil {
		return
	}

	return deleteNamespaceResouce(statefulSet, config.ClusterName(), projectName, namespace, "stateful-set", name, config.DryRun())
}

func deleteProjectResouce(resource client.ResourceClient, clusterName, projectName, kind, name string, dryRun bool) (deleted bool, err error) {
	if exists, err := resource.Exists(); err != nil || !exists {
		if err != nil {
			return false, err
		}
		logrus.
			WithField("project-name", projectName).
			WithField("resouce-name", name).
			WithField("cluster-name", clusterName).
			Infof("No %s skip delete", kind)
		return false, nil
	}

	deleted, err = resource.Delete(dryRun)
	return
}

func deleteNamespaceResouce(resource client.ResourceClient, clusterName, projectName, namespace, kind, name string, dryRun bool) (deleted bool, err error) {
	if exists, err := resource.Exists(); err != nil || !exists {
		if err != nil {
			return false, err
		}
		logrus.
			WithField("project-name", projectName).
			WithField("namespace", namespace).
			WithField("resouce-name", name).
			WithField("cluster-name", clusterName).
			Infof("No %s skip delete", kind)
		return false, nil
	}

	deleted, err = resource.Delete(dryRun)
	return
}
