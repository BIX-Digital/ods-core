package ods_setup

import (
    "github.com/opendevstack/ods-core/tests/utils"
    imageClientV1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
    projectClientV1 "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    rbacv1client "k8s.io/client-go/kubernetes/typed/rbac/v1"
    "path"
    "runtime"
    "testing"
    "time"
)

func TestCreateOdsProject(t *testing.T) {
    namespace := "myods"
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
    rbacV1Client, err := rbacv1client.NewForConfig(config)
    if err != nil {
        t.Fatalf("Cannot initialize RBAC Client: %s", err)
    }

    roleBindings, _ := rbacV1Client.RoleBindings(namespace).List(metav1.ListOptions{})

    if err = utils.FindRoleBinding(roleBindings, "system:authenticated", "Group", "", "view"); err != nil {
        t.Fatal(err)
    }

    if err = utils.FindRoleBinding(roleBindings, "system:authenticated", "Group", "", "view"); err != nil {
        t.Fatal(err)
    }

    clusterRoleBindings, _ := rbacV1Client.ClusterRoleBindings().List(metav1.ListOptions{})
    if err = utils.FindClusterRoleBinding(clusterRoleBindings, "system:authenticated", "Group", "", "system:image-puller"); err != nil {
        t.Fatal(err)
    }

    if err = utils.FindClusterRoleBinding(clusterRoleBindings, "deployment", "ServiceAccount", namespace, "cluster-admin"); err != nil {
        t.Fatal(err)
    }

    _, filename, _, _ := runtime.Caller(0)
    dir := path.Join(path.Dir(filename), "..", "..", "ods-setup", "ocp-config", "cd-user")
    stdout, stderr, err = utils.RunCommandWithWorkDir("tailor", []string{"status", "--force", "--reveal-secrets", "-n", namespace}, dir)
    if err != nil {
        t.Fatalf(
            "Execution of tailor failed: \nStdOut: %s\nStdErr: %s",
            stdout,
            stderr)
    }

}

func TestCreateDefaultOdsProject(t *testing.T) {
    namespace := "cd"
    _ = utils.RemoveProject(namespace)
    stdout, stderr, err := utils.RunScriptFromBaseDir("ods-setup/setup-ods-project.sh", []string{
        "--force",
        "--verbose",
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

}
