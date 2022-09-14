# Contributing info

Some style guidelines in
[radian-software/contributor-guide](https://github.com/radian-software/contributor-guide).

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
