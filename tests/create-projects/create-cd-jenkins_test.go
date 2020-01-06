package create_projects

import (
	"encoding/base64"
	"fmt"
	"github.com/opendevstack/ods-core/tests/utils"
	"path"
	"runtime"
	"testing"
)

func TestCreateJenkinsWithOutProjectId(t *testing.T) {
	stdout, stderr, err := utils.RunScriptFromBaseDir("create-projects/create-cd-jenkins.sh", []string{"--force", "--verbose"})
	if err == nil {
		t.Fatalf(
			"Execution of `create-cd-jenkins.sh` must fail if no PROJECT_ID is set: \nStdOut: %s\nStdErr: %s",
			stdout,
			stderr)
	}
}

func TestCreateJenkinsWithOutCDUserType(t *testing.T) {
	user := base64.StdEncoding.EncodeToString([]byte("myuser"))
	secret := base64.StdEncoding.EncodeToString([]byte("mysecret"))
	stdout, stderr, err := utils.RunScriptFromBaseDir("create-projects/create-cd-jenkins.sh", []string{"--force", "--verbose"},
		utils.PROJECT_ENV_VAR,
		// "CD_USER_TYPE=general",
		fmt.Sprintf("CD_USER_ID_B64=%s", user),
		fmt.Sprintf("PIPELINE_TRIGGER_SECRET=%s", secret))
	if err == nil {
		t.Fatalf(
			"Execution of `create-cd-jenkins.sh` must fail if no PROJECT_ID is set: \nStdOut: %s\nStdErr: %s",
			stdout,
			stderr)
	}
}

func TestCreateJenkinsWithOutSecret(t *testing.T) {
	user := base64.StdEncoding.EncodeToString([]byte("myuser"))
	//secret := base64.StdEncoding.EncodeToString([]byte("mysecret"))
	stdout, stderr, err := utils.RunScriptFromBaseDir("create-projects/create-cd-jenkins.sh", []string{"--force", "--verbose"},
		utils.PROJECT_ENV_VAR,
		"CD_USER_TYPE=general",
		fmt.Sprintf("CD_USER_ID_B64=%s", user))
	// fmt.Sprintf("PIPELINE_TRIGGER_SECRET=%s",secret)	)
	if err == nil {
		t.Fatalf(
			"Execution of `create-cd-jenkins.sh` must fail if no PROJECT_ID is set: \nStdOut: %s\nStdErr: %s",
			stdout,
			stderr)
	}
}

func TestCreateJenkins(t *testing.T) {
	_ = utils.RemoveAllTestOCProjects()
	odsNamespace := "ods"
	stdout, stderr, err := utils.RunScriptFromBaseDir("create-projects/create-projects.sh", []string{}, utils.PROJECT_ENV_VAR)
	if err != nil {
		t.Fatalf(
			"Execution of `create-project.sh` failed: \nStdOut: %s\nStdErr: %s",
			stdout,
			stderr)
	}

	values, err := utils.ReadValues()
	if err != nil {
		t.Fatalf(
			"Could not read ods-core.env")
	}

	user := values["CD_USER_ID_B64"]
	secret := values["PIPELINE_TRIGGER_SECRET_B64"]

	stdout, stderr, err = utils.RunScriptFromBaseDir("create-projects/create-cd-jenkins.sh", []string{"--force", "--verbose", "--ods-namespace", odsNamespace},
		utils.PROJECT_ENV_VAR,
		"CD_USER_TYPE=general",
		fmt.Sprintf("CD_USER_ID_B64=%s", user),
		fmt.Sprintf("PIPELINE_TRIGGER_SECRET=%s", secret))
	if err != nil {
		t.Fatalf(
			"Execution of `create-cd-jenkins.sh` failed: \nStdOut: %s\nStdErr: %s",
			stdout,
			stderr)
	}
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "..", "..", "create-projects", "ocp-config", "cd-jenkins")

	stdout, stderr, err = utils.RunCommandWithWorkDir("tailor", []string{"status", "--force", "--reveal-secrets", "-n", utils.PROJECT_NAME_CD,
		fmt.Sprintf("--param=PROJECT=%s", utils.PROJECT_NAME),
		fmt.Sprintf("--param=CD_USER_ID_B64=%s", user),
		fmt.Sprintf("--param=NAMESPACE=%s", odsNamespace),
		"--selector", "template=cd-jenkins-template",
		fmt.Sprintf("--param=%s", fmt.Sprintf("PROXY_TRIGGER_SECRET_B64=%s", secret))}, dir)
	if err != nil {

		t.Fatalf(
			"Execution of tailor failed: \nStdOut: %s\nStdErr: %s",
			stdout,
			stderr)
	}
}
