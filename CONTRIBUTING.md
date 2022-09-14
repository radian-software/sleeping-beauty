# Contributing info

Some style guidelines in
[radian-software/contributor-guide](https://github.com/radian-software/contributor-guide).

## Releasing a new version

* [Install GoReleaser](https://goreleaser.com/install/)
* Update the changelog to have the new version at the top (replace
  `Unreleased` with a version number)
* Run `make releasenotes` to check you updated the changelog correctly
* Run `make release` to do all the publishing steps
