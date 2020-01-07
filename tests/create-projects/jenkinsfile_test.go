package create_projects

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/opendevstack/ods-core/tests/utils"
	v1 "github.com/openshift/api/build/v1"
	buildClientV1 "github.com/openshift/client-go/build/clientset/versioned/typed/build/v1"
	projectClientV1 "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rbacv1client "k8s.io/client-go/kubernetes/typed/rbac/v1"
	"net/http"
	"path"
	"runtime"
	"testing"
	"time"
)

func TestCreateProjectWithJenkinsFile(t *testing.T) {
	_ = utils.RemoveAllTestOCProjects()

	values, err := utils.ReadValues()
	if err != nil {
		t.Fatalf("Error reading ods-core.env: %s", err)
	}

	request := utils.RequestBuild{
		Repository: "ods-core",
		Branch:     "ci/cd",
		Project:    "opendevstack",
		Env: []utils.EnvPair{
			{
				Name:  "PROJECT_ID",
				Value: utils.PROJECT_NAME,
			},
			{
				Name:  "CD_USER_TYPE",
				Value: "general",
			},
			{
				Name:  "CD_USER_ID_B64",
				Value: values["CD_USER_ID_B64"],
			},
			{
				Name:  "PIPELINE_TRIGGER_SECRET",
				Value: values["PIPELINE_TRIGGER_SECRET_B64"],
			},
			{
				Name:  "ODS_GIT_REF",
				Value: "ci/cd",
			},
		},
	}

	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Could not marchal json: %s", err)
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	_, err = http.Post(
		fmt.Sprintf("https://webhook-proxy-prov-cd.172.17.0.1.nip.io/build?trigger_secret=%s&jenkinsfile_path=create-projects/Jenkinsfile&component=ods-corejob-create-project-%s",
			values["PIPELINE_TRIGGER_SECRET"],
			utils.PROJECT_NAME),
		"application/json",
		bytes.NewBuffer(body))

	if err != nil {
		t.Fatalf("Could not post request: %s", err)
	}

	config, err := utils.GetOCClient()
	if err != nil {
		t.Fatalf("Error creating OC config: %s", err)
	}

	buildClient, err := buildClientV1.NewForConfig(config)
	if err != nil {
		t.Fatalf("Error creating Build client: %s", err)
	}
	time.Sleep(10 * time.Second)
	build, err := buildClient.Builds("prov-cd").Get(fmt.Sprintf("ods-corejob-create-project-%s-ci-cd-1", utils.PROJECT_NAME), metav1.GetOptions{})
	count := 0
	max := 240
	for (err != nil || build.Status.Phase == v1.BuildPhaseNew || build.Status.Phase == v1.BuildPhasePending || build.Status.Phase == v1.BuildPhaseRunning) && count < max {
		build, err = buildClient.Builds("prov-cd").Get(fmt.Sprintf("ods-corejob-create-project-%s-ci-cd-1", utils.PROJECT_NAME), metav1.GetOptions{})
		time.Sleep(2 * time.Second)
		if err != nil {
			t.Log("Build is still not available")
		} else {
			t.Logf("Waiting for build. Current status: %s", build.Status.Phase)
		}
		count++
	}

	if count >= max || build.Status.Phase != v1.BuildPhaseComplete {
		stdout, stderr, _ := utils.RunScriptFromBaseDir(
			"tests/scripts/utils/print-jenkins-log.sh",
			[]string{fmt.Sprintf("ods-corejob-create-project-%s-ci-cd-1", utils.PROJECT_NAME)})
		if count >= max {
			t.Fatalf(
				"Timeout during build: \nStdOut: %s\nStdErr: %s",
				stdout,
				stderr)
		} else {
			t.Fatalf(
				"Error during build: \nStdOut: %s\nStdErr: %s",
				stdout,
				stderr)
		}

	}
	client, err := projectClientV1.NewForConfig(config)
	if err != nil {
		t.Fatalf("Error creating Project client: %s", err)
	}
	projects, err := client.Projects().List(metav1.ListOptions{})
	if err != nil {
		t.Fatalf("Cannot list projects: %s", err)
	}

	if err = utils.FindProject(projects, utils.PROJECT_NAME_CD); err != nil {
		t.Fatal(err)
	}
	if err = utils.FindProject(projects, utils.PROJECT_NAME_TEST); err != nil {
		t.Fatal(err)
	}
	if err = utils.FindProject(projects, utils.PROJECT_NAME_DEV); err != nil {
		t.Fatal(err)
	}

	rbacV1Client, err := rbacv1client.NewForConfig(config)
	if err != nil {
		t.Fatalf("Cannot initialize RBAC Client: %s", err)
	}
	roleBindings, _ := rbacV1Client.RoleBindings(utils.PROJECT_NAME_CD).List(metav1.ListOptions{})

	if err = utils.FindRoleBinding(roleBindings, "jenkins", "ServiceAccount", utils.PROJECT_NAME_CD, "edit"); err != nil {
		t.Fatal(err)
	}
	if err = utils.FindRoleBinding(roleBindings, "default", "ServiceAccount", utils.PROJECT_NAME_CD, "edit"); err != nil {
		t.Fatal(err)
	}

	if err = utils.FindRoleBinding(roleBindings, fmt.Sprintf("system:serviceaccounts:%s", utils.PROJECT_NAME_DEV), "Group", "", "system:image-puller"); err != nil {
		t.Fatal(err)
	}

	if err = utils.FindRoleBinding(roleBindings, fmt.Sprintf("system:serviceaccounts:%s", utils.PROJECT_NAME_TEST), "Group", "", "system:image-puller"); err != nil {
		t.Fatal(err)
	}

	roleBindings, _ = rbacV1Client.RoleBindings(utils.PROJECT_NAME_DEV).List(metav1.ListOptions{})
	if err = utils.FindRoleBinding(roleBindings, "default", "ServiceAccount", utils.PROJECT_NAME_DEV, "system:image-builder"); err != nil {
		t.Fatal(err)
	}

	if err = utils.FindRoleBinding(roleBindings, fmt.Sprintf("system:serviceaccounts:%s", utils.PROJECT_NAME_TEST), "Group", "", "system:image-puller"); err != nil {
		t.Fatal(err)
	}

	if err = utils.FindRoleBinding(roleBindings, "jenkins", "ServiceAccount", utils.PROJECT_NAME_CD, "admin"); err != nil {
		t.Fatal(err)
	}

	roleBindings, _ = rbacV1Client.RoleBindings(utils.PROJECT_NAME_TEST).List(metav1.ListOptions{})
	if err = utils.FindRoleBinding(roleBindings, "default", "ServiceAccount", utils.PROJECT_NAME_TEST, "system:image-builder"); err != nil {
		t.Fatal(err)
	}

	if err = utils.FindRoleBinding(roleBindings, "jenkins", "ServiceAccount", utils.PROJECT_NAME_CD, "admin"); err != nil {
		t.Fatal(err)
	}

	t.Log("WARNING: Seeding special and default permission groups is not tested yet!")

	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "..", "..", "create-projects", "ocp-config", "cd-jenkins")

	user := values["CD_USER_ID_B64"]
	secret := values["PIPELINE_TRIGGER_SECRET_B64"]

	stdout, stderr, err := utils.RunCommandWithWorkDir("tailor", []string{"status", "--force", "--reveal-secrets", "-n", utils.PROJECT_NAME_CD,
		fmt.Sprintf("--param=PROJECT=%s", utils.PROJECT_NAME),
		fmt.Sprintf("--param=CD_USER_ID_B64=%s", user),
		"--selector", "template=cd-jenkins-template",
		fmt.Sprintf("--param=%s", fmt.Sprintf("PROXY_TRIGGER_SECRET_B64=%s", secret))}, dir)
	if err != nil {

		t.Fatalf(
			"Execution of tailor failed: \nStdOut: %s\nStdErr: %s",
			stdout,
			stderr)
	}

}
