FILES_TO_FMT=$(shell find . -name '*.go' -print)

clean-docs:
	@rm -rf ./docs/_intermediate

clean: clean-docs

update-docs: intermediate-docs embedmd
	@./scripts/replace-rulenames-with-doclinks.sh
	@embedmd -w ./docs/index.md

intermediate-docs:
	@mkdir -p ./docs/_intermediate
	@go run ./main.go -h > ./docs/_intermediate/help.txt
	@go run ./main.go completion -h > ./docs/_intermediate/completion.txt
	@go run ./main.go lint -h > ./docs/_intermediate/lint.txt
	@go run ./main.go rules > ./docs/_intermediate/rules.txt
	@echo "Can't automate everything, please replace the #Rules section of index.md with the contents of ./docs/_intermediate/rules.txt"

embedmd:
	@go install github.com/campoy/embedmd@v1.0.0

.PHONY: fmt check-fmt
fmt:
	@gofmt -s -w $(FILES_TO_FMT)
	@goimports -w $(FILES_TO_FMT)

check-fmt: fmt
	@git diff --exit-code -- $(FILES_TO_FMT)

.PHONY: test
test:
	@go test ./...

.PHONY: lint
lint:
	@echo "Running golangci-lint"
	golangci-lint run

.PHONY: check
check: test lint
