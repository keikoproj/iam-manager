# Developing


### Requirements
* Go1.13
* Docker
* Your favorite IDE
* kubebuilder

### Quickstart

First, go and fork the Github repo to your own personal project. Once that's
done, set up a local build environment off of the original Github repo. Then we
add in your fork'ed repo as a new target for doing git pushes.

    $ go clean -modcache
    $ go get -v github.com/keikoproj/iam-manager
    $ cd "$(go env GOPATH)/src/github.com/keikoproj/iam-manager"
    $ make test
    $ go mod vendor
    $ git remote add myfork <your fork>

### Install Kubebuilder

Kubebuilder is a requirement for the testsuite.. you can install it quickly
on your own or with our make target:

    $ make kubebuilder

### Build project

    $ make

### Running Tests

There are several environment variables that must be set in order for the
test suite to work. The [Makefile](/Makefile) sets these for you, so please
use the `make test` target to run tests:

    $ make test
