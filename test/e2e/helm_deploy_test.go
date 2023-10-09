package e2e

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opendevstack/ods-pipeline/pkg/artifact"
	ott "github.com/opendevstack/ods-pipeline/pkg/odstasktest"
	"github.com/opendevstack/ods-pipeline/pkg/pipelinectxt"
	ttr "github.com/opendevstack/ods-pipeline/pkg/tektontaskrun"
	tekton "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/yaml"
)

const (
	localRegistry = "localhost:5000"
	kindRegistry  = "ods-pipeline-registry.kind:5000"
)

type imageImportParams struct {
	externalRef string
	namespace   string
	workdir     string
}

func TestDeployHelmSkipsWithoutNamespaceTask(t *testing.T) {
	if err := runTask(
		// no "namespace" param set
		ott.WithGitSourceWorkspace(t, "../testdata/workspaces/helm-sample-app", namespaceConfig.Name),
	); err != nil {
		t.Fatal(err)
	}
}

func TestDeployHelmInstallsIntoSeparateNamespace(t *testing.T) {
	k8sClient := newK8sClient(t)
	releaseNamespace, nsCleanup := createReleaseNamespaceOrFatal(t, k8sClient, namespaceConfig.Name)
	defer nsCleanup()

	if err := runTask(
		ttr.WithStringParams(
			map[string]string{"namespace": releaseNamespace.Name},
		),
		ott.WithGitSourceWorkspace(t, "../testdata/workspaces/helm-sample-app", namespaceConfig.Name),
		importImage(t, "index.docker.io/crccheck/hello-world"),
		createSampleAppPrivateKeySecret(t, k8sClient, namespaceConfig.Name),
		ttr.AfterRun(func(config *ttr.TaskRunConfig, run *tekton.TaskRun, logs bytes.Buffer) {
			ott.AssertFileContentContains(t,
				config.WorkspaceConfigs["source"].Dir,
				filepath.Join(pipelinectxt.DeploymentsPath, fmt.Sprintf("diff-%s.txt", releaseNamespace.Name)),
				"Release was not present in Helm.  Diff will show entire contents as new.",
				"Deployment (apps) has been added",
				"Secret (v1) has been added",
				"Service (v1) has been added",
			)
			ott.AssertFileContentContains(t,
				config.WorkspaceConfigs["source"].Dir,
				filepath.Join(pipelinectxt.DeploymentsPath, fmt.Sprintf("release-%s.txt", releaseNamespace.Name)),
				"Installing it now.",
				fmt.Sprintf("NAMESPACE: %s", releaseNamespace.Name),
				"STATUS: deployed",
				"REVISION: 1",
			)
		}),
	); err != nil {
		t.Fatal(err)
	}
}

func newK8sClient(t *testing.T) *kubernetes.Clientset {
	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		t.Fatal(err)
	}
	kubernetesClientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		t.Fatal(err)
	}
	return kubernetesClientset
}

func createSampleAppPrivateKeySecret(t *testing.T, clientset *kubernetes.Clientset, namespace string) ttr.TaskRunOpt {
	return func(c *ttr.TaskRunConfig) error {
		secret, err := readPrivateKeySecret()
		if err != nil {
			return err
		}
		t.Logf("Creating secret %q in namespace %q ...", secret.Name, namespace)
		_, err = clientset.CoreV1().
			Secrets(namespace).
			Create(context.TODO(), secret, metav1.CreateOptions{})
		return err
	}
}

func readPrivateKeySecret() (*corev1.Secret, error) {
	bytes, err := os.ReadFile("../testdata/fixtures/tasks/secret.yaml")
	if err != nil {
		return nil, err
	}

	var secretSpec corev1.Secret
	err = yaml.Unmarshal(bytes, &secretSpec)
	if err != nil {
		return nil, err
	}
	return &secretSpec, nil
}

func createReleaseNamespaceOrFatal(t *testing.T, clientset *kubernetes.Clientset, ctxtNamespace string) (namespace *corev1.Namespace, cleanup func()) {
	namespace, err := createReleaseNamespace(clientset, ctxtNamespace, "foobar") //fix name
	if err != nil {
		t.Fatal(err)
	}
	return namespace, func() {
		if err := clientset.CoreV1().Namespaces().Delete(context.TODO(), namespace.Name, metav1.DeleteOptions{}); err != nil {
			t.Logf("Failed to delete namespace %s: %s", namespace.Name, err)
		}
	}
}

func createReleaseNamespace(clientset *kubernetes.Clientset, ctxtNamespace, name string) (*corev1.Namespace, error) {
	ns, err := clientset.CoreV1().Namespaces().Create(
		context.TODO(),
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}},
		metav1.CreateOptions{},
	)
	if err != nil {
		return nil, err
	}
	_, err = clientset.RbacV1().RoleBindings(name).Create(
		context.Background(),
		&rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pipeline-deployer",
				Namespace: name,
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "pipeline",
					Namespace: ctxtNamespace,
				},
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     "edit",
			},
		},
		metav1.CreateOptions{})

	return ns, err
}

// 	resourceName := fmt.Sprintf("%s-%s", ctxt.ODS.Component, "helm-sample-app")
// 	_, err := checkService(ctxt.Clients.KubernetesClientSet, separateReleaseNamespace, resourceName)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	_, err = checkDeployment(ctxt.Clients.KubernetesClientSet, separateReleaseNamespace, resourceName)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Verify log output massaging
// 	doNotWantLogMsg := "plugin \"diff\" exited with error"
// 	if strings.Contains(string(ctxt.CollectedLogs), doNotWantLogMsg) {
// 		t.Fatalf("Do not want:\n%s\n\nGot:\n%s", doNotWantLogMsg, string(ctxt.CollectedLogs))
// 	}
// 	wantLogMsg := "identified at least one change"
// 	if !strings.Contains(string(ctxt.CollectedLogs), wantLogMsg) {
// 		t.Fatalf("Want:\n%s\n\nGot:\n%s", wantLogMsg, string(ctxt.CollectedLogs))
// 	}
// },

func importImage(t *testing.T, externalRef string) ttr.TaskRunOpt {
	return func(c *ttr.TaskRunConfig) error {
		iip := imageImportParams{
			externalRef: externalRef,
			namespace:   namespaceConfig.Name,
			workdir:     c.WorkspaceConfigs["source"].Dir,
		}

		var err error
		cmds := [][]string{
			{"pull", iip.externalRef},
			{"tag", iip.externalRef, iip.internalRef(localRegistry)},
			{"push", iip.internalRef(localRegistry)},
		}
		for _, args := range cmds {
			if err == nil {
				_, _, err = runBuffered("docker", args)
			}
		}
		if err != nil {
			return fmt.Errorf("docker cmd failed: %s", err)
		}

		err = pipelinectxt.WriteJsonArtifact(artifact.Image{
			Ref:        iip.internalRef(kindRegistry),
			Registry:   kindRegistry,
			Repository: iip.namespace,
			Name:       iip.name(),
			Tag:        "latest",
			Digest:     "not needed",
		}, filepath.Join(iip.workdir, pipelinectxt.ImageDigestsPath), fmt.Sprintf("%s.json", iip.name()))
		if err != nil {
			return fmt.Errorf("failed to write artifact: %s", err)
		}

		t.Log("Imported image", iip.internalRef(localRegistry))

		return nil
	}

}

func checkIfImageExists(t *testing.T, name string) bool {
	t.Helper()
	_, _, err := runBuffered("docker", []string{"inspect", name})
	return err == nil
}

func (iip imageImportParams) name() string {
	parts := strings.Split(iip.externalRef, "/")
	return parts[2]
}

func (iip imageImportParams) internalRef(registry string) string {
	parts := strings.Split(iip.externalRef, "/")
	return fmt.Sprintf("%s/%s/%s", registry, iip.namespace, parts[2])
}

func runBuffered(executable string, args []string) (outBytes, errBytes []byte, err error) {
	cmd := exec.Command(executable, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	outBytes = stdout.Bytes()
	errBytes = stderr.Bytes()
	return outBytes, errBytes, err
}
