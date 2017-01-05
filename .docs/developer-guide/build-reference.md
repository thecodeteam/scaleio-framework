# Build Reference

How to build the ScaleIO-Framework

---

## Build Requirements
This project has very few build requirements, but there are still one or two
items of which to be aware. Also, please note that these are the requirements to
*build* `ScaleIO-Framework`, not run it.

Requirement | Version
------------|--------
Operating System | Linux, OSX
[Go](https://golang.org/) | >=1.7.3

## Cross-Compilation
This project only currently supports running on Linux based platforms. Specifically
those outlined in the Requirements section on the main landing (aka README.md)
page. Although you can develop on OSX you will not be able to run on the OSX
platform.

## Performing Builds
Building from source is pretty simple as all steps follow traditional golang
based projects. After forking the github project, there are two components that
make up the Framework, the scheduler and the executor. Simply navigate to each
directory and run the following build command:
`glide up && GOOS=linux GOARCH=amd64 go build .`

The output will look similar to the following:

```sh
[INFO] Downloading dependencies. Please wait...
[INFO] Fetching updates for github.com/stretchr/testify.
[INFO] Fetching updates for github.com/Sirupsen/logrus.
[INFO] Fetching updates for github.com/dvonthenen/goxplatform.
[INFO] Setting version for github.com/stretchr/testify to v1.1.3.
[INFO] Setting version for github.com/Sirupsen/logrus to v0.10.0.
[INFO] Resolving imports
[INFO] Fetching golang.org/x/sys/unix into /Users/vonthd/go/src/github.com/codedellemc/scaleio-framework/scaleio-scheduler/vendor
[INFO] Fetching github.com/golang/protobuf/proto into /Users/vonthd/go/src/github.com/codedellemc/scaleio-framework/scaleio-scheduler/vendor
[INFO] Fetching github.com/gogo/protobuf/jsonpb into /Users/vonthd/go/src/github.com/codedellemc/scaleio-framework/scaleio-scheduler/vendor
[INFO] Fetching github.com/codegangsta/negroni into /Users/vonthd/go/src/github.com/codedellemc/scaleio-framework/scaleio-scheduler/vendor
[INFO] Fetching github.com/gorilla/mux into /Users/vonthd/go/src/github.com/codedellemc/scaleio-framework/scaleio-scheduler/vendor
[INFO] Fetching github.com/twinj/uuid into /Users/vonthd/go/src/github.com/codedellemc/scaleio-framework/scaleio-scheduler/vendor
[INFO] Fetching github.com/gorilla/context into /Users/vonthd/go/src/github.com/codedellemc/scaleio-framework/scaleio-scheduler/vendor
[INFO] Downloading dependencies. Please wait...
[INFO] Setting references for remaining imports
[INFO] Project relies on 10 dependencies.
```

Upon completion of a successful build, you will not receive a "positive" notification
of that build, but rather you will find a completed binary at the root of the
folder.

## Making modifications to source and building
If you plan on making changes to the source for developing or testing purposes,
you must follow a couple of prerequisites due to the layout of the project in
GitHub. Since there exists two projects in in the same repo in which one project,
the executor, depends on another, the scheduler, there currently is an issue with
glide when pulling those dependencies for each project. To remedy the issue for
private builds, you must:

1. Fork the project
2. Clone your fork into your own workspace
3. Create a git branch, make your changes and push your branch to your private repo
4. Run the ./switch.sh file which swaps the current glide.yaml with glide.yaml.dev
5. Then open the glide.yaml and replace the repo and ref properties of the scaleio-framework package with your fork and the commit you want to build against.
  ```
  - package: github.com/codedellemc/scaleio-framework
    ref:     a8a1be0c946a19f97fdc962150626186f3315078
    vcs:     git
    repo:    https://github.com/dvonthenen/scaleio-framework
  ```
6. Then run ./build.sh. If the glide step fails, re-run the ./build.sh command.

## Version File
There is a file at the root of the project named `VERSION`. The file contains
a single line with the *target* version of the project in the file. The version
follows the format:

  `(?<major>\d+)\.(?<minor>\d+)\.(?<patch>\d+)(-rc\d+)?`

For example, during active development of version `0.1.0` the file would
contain the version `0.1.0`. When it's time to create `0.4.0`'s first
release candidate the version in the file will be changed to `0.1.0-rc1`. And
when it's time to release `0.1.0` the version is changed back to `0.1.0`.

Please note that we've discussed making the actively developed version the
targeted version with a `-dev` suffix, but trying this resulted in confusion
for the RPM and DEB package managers when using `unstable` releases.

So what's the point of the file if it's basically duplicating the utility of a
tag? Well, the `VERSION` file in fact has two purposes:

  1. First and foremost updating the `VERSION` file with the same value as that
     of the tag used to create a release provides a single, contextual reason to
     push a commit and tag. Otherwise some random commit off of `master` would
     be tagged as a release candidate or release. Always using the commit that
     is related to updating the `VERSION` file is much cleaner.

  2. The contents of the `VERSION` file are also used during the build process
     as a means of overriding the output of a `git describe`. This enables the
     semantic version injected into the produced binary to be created using
     the *targeted* version of the next release and not just the value of the
     last, tagged commit.
