VERSION := $(shell cat CHANGELOG.md | grep '^##' | head -n1 | tr -d '# ')

# Remove blank lines from beginning and end of file
# https://unix.stackexchange.com/a/552198
REMOVE_BLANK_LINES = awk 'NF {p=1} p' | tac | awk 'NF {p=1} p' | tac

# Shell command to strip text before first heading in changelog, then
# drop first line of file, then strip next heading and following text.
# Just copied from stackoverflow, if it stops working, then find
# another one.
RELEASE_NOTES = cat CHANGELOG.md | sed '/^\#\#/,$$!d' | tail -n+2 | sed -n '/^\#\#/q;p' | $(REMOVE_BLANK_LINES)

.PHONY: build
build:
	go build ./cmd/sleepingd

.PHONY: test-unit
test-unit:
	go test ./lib/sleepingd $(TEST_FLAGS)

.PHONY: test-integration
test-integration: build
	./docker/run_in_docker.bash sleeping-beauty-integration-test:latest \
		./test/integration/run.bash $(TEST_FLAGS)

.PHONY: test
test: test-unit test-integration

.PHONY: version
version:
	@echo "Current version is $(VERSION) according to CHANGELOG.md"

.PHONY: releasenotes
releasenotes: version
	@printf '------------------------------\n'
	@$(RELEASE_NOTES)
	@printf '------------------------------\n'

.PHONY: release
release:
	@echo "Releasing version $(VERSION)"
	@$(RELEASE_NOTES) > .releasenotes.tmp.md
	git tag v$(VERSION) HEAD $(if $(FORCE),-f,)
	git push origin v$(VERSION) $(if $(FORCE),-f,)
	goreleaser release --rm-dist --release-notes=.releasenotes.tmp.md
