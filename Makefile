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

test:
	@go test ./...

