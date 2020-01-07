package ods_setup

import (
	"github.com/opendevstack/ods-core/tests/utils"
	imageClientV1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	projectClientV1 "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestCreateOdsProject(t *testing.T) {
	namespace := "ods"
	_ = utils.RemoveProject(namespace)
	stdout, stderr, err := utils.RunScriptFromBaseDir("ods-setup/setup-ods-project.sh", []string{
		"--force",
		"--verbose",
		"--namespace",
		namespace,
	})
	if err != nil {
		t.Fatalf(
			"Execution of `setup-ods-project.sh` failed: \nStdOut: %s\nStdErr: %s",
			stdout,
			stderr)
	}

	config, err := utils.GetOCClient()
	if err != nil {
		t.Fatalf("Error creating OC config: %s", err)
	}
	client, err := projectClientV1.NewForConfig(config)
	if err != nil {
		t.Fatalf("Error creating Project client: %s", err)
	}
	projects, err := client.Projects().List(metav1.ListOptions{})
	if err != nil {
		t.Fatalf("Cannot list projects: %s", err)
	}

	if err = utils.FindProject(projects, namespace); err != nil {
		t.Fatal(err)
	}

	gitReference := "production"

	stdout, stderr, err = utils.RunScriptFromBaseDir("tests/scripts/deploy-mocks.sh", []string{
		"--verbose",
	})
	if err != nil {
		t.Fatalf(
			"Execution of `deploy-mocks.sh` failed: \nStdOut: %s\nStdErr: %s",
			stdout,
			stderr)
	}
	time.Sleep(5 * time.Second)

	stdout, stderr, err = utils.RunScriptFromBaseDir("tests/scripts/setup-mocked-ods-repo.sh", []string{
		"--verbose",
		"--ods-ref", gitReference,
	})
	if err != nil {
		t.Fatalf(
			"Execution of `setup-mocked-ods-repo.sh` failed: \nStdOut: %s\nStdErr: %s",
			stdout,
			stderr)
	}

	stdout, stderr, err = utils.RunScriptFromBaseDir("ods-setup/setup-jenkins-images.sh", []string{
		"--verbose",
		"--force",
		"--ods-ref", gitReference,
		"--namespace", namespace,
	})
	if err != nil {
		t.Fatalf(
			"Execution of `setup-mocked-ods-repo.sh` failed: \nStdOut: %s\nStdErr: %s",
			stdout,
			stderr)
	}

	imageClient, err := imageClientV1.NewForConfig(config)
	if err != nil {
		t.Fatalf("Error creating Image client: %s", err)
	}

	images, err := imageClient.ImageStreams(namespace).List(metav1.ListOptions{})
	if err = utils.FindImageTag(images, "jenkins-master", "test"); err != nil {
		t.Fatal(err)
	}

	if err = utils.FindImageTag(images, "jenkins-slave-base", "test"); err != nil {
		t.Fatal(err)
	}

	if err = utils.FindImageTag(images, "jenkins-webhook-proxy", "test"); err != nil {
		t.Fatal(err)
	}

	if err = utils.FindImageTag(images, "jenkins-master", "latest"); err != nil {
		t.Fatal(err)
	}

	if err = utils.FindImageTag(images, "jenkins-slave-base", "latest"); err != nil {
		t.Fatal(err)
	}

	if err = utils.FindImageTag(images, "jenkins-webhook-proxy", "latest"); err != nil {
		t.Fatal(err)
	}

}
