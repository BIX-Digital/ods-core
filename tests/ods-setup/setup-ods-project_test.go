package ods_setup

import (
	"github.com/opendevstack/ods-core/tests/utils"
	imageClientV1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	projectClientV1 "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
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

	gitReference := "cicdtests"

	stdout, stderr, err = utils.RunScriptFromBaseDir("ods-setup/setup-jenkins-images.sh", []string{
		"--verbose",
		"--force",
		"--ods-ref", gitReference,
		"--namespace", namespace,
	})
	if err != nil {
		t.Fatalf(
			"Execution of `setup-jenkins-images.sh` failed: \nStdOut: %s\nStdErr: %s",
			stdout,
			stderr)
	}

	imageClient, err := imageClientV1.NewForConfig(config)
	if err != nil {
		t.Fatalf("Error creating Image client: %s", err)
	}

	images, err := imageClient.ImageStreams(namespace).List(metav1.ListOptions{})
	if err = utils.FindImageTag(images, "jenkins-master", gitReference); err != nil {
		t.Fatalf("%s\nScript:\nStdOut: %s\nStdErr: %s",err, stdout, stderr)
	}

	if err = utils.FindImageTag(images, "jenkins-slave-base", gitReference); err != nil {
		t.Fatalf("%s\nScript:\nStdOut: %s\nStdErr: %s",err, stdout, stderr)
	}

	if err = utils.FindImageTag(images, "jenkins-webhook-proxy", gitReference); err != nil {
		t.Fatalf("%s\nScript:\nStdOut: %s\nStdErr: %s",err, stdout, stderr)
	}	
}
