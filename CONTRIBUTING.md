# Contributing info

Some style guidelines in
[radian-software/contributor-guide](https://github.com/radian-software/contributor-guide).

## Running the tests locally

```
docker build . -f ./test/integration/Dockerfile -t sleeping-beauty-integration-test
make test-unit TEST_FLAGS=-v
make test-integration TEST_FLAGS=-v
```

## Releasing a new version

* Update the changelog to have the new version at the top (replace
  `Unreleased` with a version number)
* Run `make releasenotes` to check you updated the changelog correctly
* [Install GoReleaser](https://goreleaser.com/install/)
* Export `GITHUB_TOKEN` to a personal access token with `repo` scope.
  You need **@raxod502** to add you as a collaborator on this
  repository
* Login to Docker Hub (`docker login -u yourname`). You need
  **@raxod502** to add you to the radiansoftware organization with
  permission to publish to the `sleeping-beauty` repository
* Run `make release` to do all the publishing steps

## Updating the CI image

```
GITHUB_TOKEN="..." ./test/integration/update_image.bash
```

The `radian-sb-bot` account can be used to get a properly scoped
personal access token.
