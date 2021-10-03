fmt:
	go fmt ./pkg/... ./cmd/...

vet:
	go vet ./pkg/... ./cmd/...

test:
	go test ./pkg/... ./cmd/... -coverprofile cover.out

bin: fmt vet test
	go build -o bin/kubectl-appsversion github.com/icyd/kubectl-apps-version/cmd/plugin

.PHONY: test fmt vet bin
