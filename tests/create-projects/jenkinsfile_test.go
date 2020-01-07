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
	"io/ioutil"
)

func TestCreateProjectWithJenkinsFile(t *testing.T) {
	projectName := "testt"
	projectNameCd :=  fmt.Sprintf("%s-cd", projectName)
	projectNameTest :=  fmt.Sprintf("%s-test", projectName)
	projectNameDev :=  fmt.Sprintf("%s-dev", projectName)
	
	_ = utils.RemoveProject(projectNameCd)
	_ = utils.RemoveProject(projectNameTest)
	_ = utils.RemoveProject(projectNameDev)

	values, err := utils.ReadValues()
	if err != nil {
		t.Fatalf("Error reading ods-core.env: %s", err)
	}

	request := utils.RequestBuild{
		Repository: "ods-core",
		Branch:     "cicdtests",
		Project:    "opendevstack",
		Env: []utils.EnvPair{
			{
				Name:  "PROJECT_ID",
				Value: projectName,
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
				Value: "cicdtests",
			},
			{
				Name:  "ODS_IMAGE_TAG",
				Value: values["ODS_IMAGE_TAG"],
			},
		},
	}

	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Could not marchal json: %s", err)
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	reponse, err := http.Post(
		fmt.Sprintf("https://webhook-proxy-prov-cd.172.17.0.1.nip.io/build?trigger_secret=%s&jenkinsfile_path=create-projects/Jenkinsfile&component=ods-corejob-create-project-%s",
			values["PIPELINE_TRIGGER_SECRET"],
			projectName),
		"application/json",
		bytes.NewBuffer(body))

	if err != nil  {
		t.Fatalf("Could not post request: %s", err)
	}
	
	if reponse.StatusCode >= http.StatusAccepted {
        bodyBytes, err := ioutil.ReadAll(reponse.Body)
        if err != nil {
            t.Fatal(err)
        }
        t.Fatalf("Could not post request: %s", string(bodyBytes))
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
	build, err := buildClient.Builds("prov-cd").Get(fmt.Sprintf("ods-corejob-create-project-%s-cicdtests-1", projectName), metav1.GetOptions{})
	count := 0
	max := 240
	for (err != nil || build.Status.Phase == v1.BuildPhaseNew || build.Status.Phase == v1.BuildPhasePending || build.Status.Phase == v1.BuildPhaseRunning) && count < max {
		build, err = buildClient.Builds("prov-cd").Get(fmt.Sprintf("ods-corejob-create-project-%s-cicdtests-1", projectName), metav1.GetOptions{})
		time.Sleep(2 * time.Second)
		if err != nil {
			t.Log("Build is still not available")
		} else {
			t.Logf("Waiting for build. Current status: %s", build.Status.Phase)
		}
		count++
	}

	stdout, stderr, _ := utils.RunScriptFromBaseDir(
		"tests/scripts/utils/print-jenkins-log.sh",
		[]string{fmt.Sprintf("ods-corejob-create-project-%s-cicdtests-1", projectName)})
	
		if count >= max || build.Status.Phase != v1.BuildPhaseComplete {
		
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

	if err = utils.FindProject(projects, projectNameCd); err != nil {
		t.Fatalf("%s\n Jenkins logs: \nStdOut: %s\nStdErr: %s",err, stdout, stderr)
	}
	if err = utils.FindProject(projects, projectNameTest); err != nil {
		t.Fatalf("%s\n Jenkins logs: \nStdOut: %s\nStdErr: %s",err, stdout, stderr)
	}
	if err = utils.FindProject(projects, projectNameDev); err != nil {
		t.Fatalf("%s\n Jenkins logs: \nStdOut: %s\nStdErr: %s",err, stdout, stderr)
	}

	rbacV1Client, err := rbacv1client.NewForConfig(config)
	if err != nil {
		t.Fatalf("Cannot initialize RBAC Client: %s", err)
	}
	roleBindings, _ := rbacV1Client.RoleBindings(projectNameCd).List(metav1.ListOptions{})

	if err = utils.FindRoleBinding(roleBindings, "jenkins", "ServiceAccount", projectNameCd, "edit"); err != nil {
		t.Fatalf("%s\n Jenkins logs: \nStdOut: %s\nStdErr: %s",err, stdout, stderr)
	}
	if err = utils.FindRoleBinding(roleBindings, "default", "ServiceAccount", projectNameCd, "edit"); err != nil {
		t.Fatalf("%s\n Jenkins logs: \nStdOut: %s\nStdErr: %s",err, stdout, stderr)
	}

	if err = utils.FindRoleBinding(roleBindings, fmt.Sprintf("system:serviceaccounts:%s", projectNameDev), "Group", "", "system:image-puller"); err != nil {
		t.Fatalf("%s\n Jenkins logs: \nStdOut: %s\nStdErr: %s",err, stdout, stderr)
	}

	if err = utils.FindRoleBinding(roleBindings, fmt.Sprintf("system:serviceaccounts:%s", projectNameTest), "Group", "", "system:image-puller"); err != nil {
		t.Fatalf("%s\n Jenkins logs: \nStdOut: %s\nStdErr: %s",err, stdout, stderr)
	}

	roleBindings, _ = rbacV1Client.RoleBindings(projectNameDev).List(metav1.ListOptions{})
	if err = utils.FindRoleBinding(roleBindings, "default", "ServiceAccount", projectNameDev, "system:image-builder"); err != nil {
		t.Fatalf("%s\n Jenkins logs: \nStdOut: %s\nStdErr: %s",err, stdout, stderr)
	}

	if err = utils.FindRoleBinding(roleBindings, fmt.Sprintf("system:serviceaccounts:%s", projectNameTest), "Group", "", "system:image-puller"); err != nil {
		t.Fatalf("%s\n Jenkins logs: \nStdOut: %s\nStdErr: %s",err, stdout, stderr)
	}

	if err = utils.FindRoleBinding(roleBindings, "jenkins", "ServiceAccount", projectNameCd, "admin"); err != nil {
		t.Fatalf("%s\n Jenkins logs: \nStdOut: %s\nStdErr: %s",err, stdout, stderr)
	}

	roleBindings, _ = rbacV1Client.RoleBindings(projectNameTest).List(metav1.ListOptions{})
	if err = utils.FindRoleBinding(roleBindings, "default", "ServiceAccount", projectNameTest, "system:image-builder"); err != nil {
		t.Fatalf("%s\n Jenkins logs: \nStdOut: %s\nStdErr: %s",err, stdout, stderr)
	}

	if err = utils.FindRoleBinding(roleBindings, "jenkins", "ServiceAccount", projectNameCd, "admin"); err != nil {
		t.Fatalf("%s\n Jenkins logs: \nStdOut: %s\nStdErr: %s",err, stdout, stderr)
	}

	t.Log("WARNING: Seeding special and default permission groups is not tested yet!")

	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "..", "..", "create-projects", "ocp-config", "cd-jenkins")

	user := values["CD_USER_ID_B64"]
	secret := values["PIPELINE_TRIGGER_SECRET_B64"]

	stdout, stderr, err = utils.RunCommandWithWorkDir("tailor", []string{"status", "--force", "--reveal-secrets", "-n", projectNameCd,
		fmt.Sprintf("--param=PROJECT=%s", projectName),
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
