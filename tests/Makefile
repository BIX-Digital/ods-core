.PHONY: test lint

test: create-projects/create-projects_test.go create-projects/create-projects_test.go go.mod go.sum
	@(go test github.com/opendevstack/ods-core/tests/create-projects -run ^(TestCreateJenkins))
	@(sleep 5)
	@(go test github.com/opendevstack/ods-core/tests/create-projects -run ^(TestCreateProject))
	@(sleep 5)
	@(go test github.com/opendevstack/ods-core/tests/create-projects -run ^(TestJenkinsFile))
	@(sleep 5)
	@(go test github.com/opendevstack/ods-core/tests/ods-setup -run ^(TestJenkinsFile))

lint:
	@(echo "Checking code ...")
	@(golangci-lint run)
