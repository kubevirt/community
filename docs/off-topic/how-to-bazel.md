# Bazel Introduction

- [Bazel introduction](http://brendanjryan.com/golang/bazel/2018/05/12/building-go-applications-with-bazel.html)
    - OpenSource software of internal Google tool called Blaze
    - Created language called Skylark/Starlark (inspired in Python)
    - Your application’s build process is guaranteed to be completely reproducible and consistent
    - Bazel also takes strides to make your builds faster, spreading work across all of your machine’s processing power
    - Agnostic Language
    - Consistent Developer Experience
    - Scaffolfind dependencies and BUILD file generation with Gazelle

- [Skylark/Starlark Language](https://docs.bazel.build/versions/master/skylark/language.html)
- Bazel Rules:
    - [Docker](https://github.com/bazelbuild/rules_docker)
    - [Webtesting](https://github.com/bazelbuild/rules_webtesting)

- gRPC:
    - [Building gRPC services with bazel and rules_protobuf](https://grpc.io/blog/bazel_rules_protobuf)

- Others:
    - https://golang.org/pkg/testing/
    - https://github.com/bazelbuild/bazel-gazelle
    - https://awesomebazel.com/
    - [Building Go with Bazel](https://www.youtube.com/watch?v=2TKxuERTnks)
    - [Building Software at Google Scale](https://www.youtube.com/watch?v=2qv3fcXW1mg)
    - [Watch and rebuild (or run tests)](https://github.com/bazelbuild/bazel-watcher)
    - [Dave Cheney - Reproducible Builds](http://go-talks.appspot.com/github.com/davecheney/presentations/reproducible-builds.slide#1)

### Setting up Bazel for Golang

- Create [WORKSPACE file](http://brendanjryan.com/golang/bazel/2018/05/12/building-go-applications-with-bazel.html#setting-up-bazel-for-go):
```
# Bazel core
load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")

# Golang Rules
git_repository(
    name = "io_bazel_rules_go",
    remote = "https://github.com/bazelbuild/rules_go.git",
    sha256 = "4442d82a001f378d0605cbbca3fb529979a1c3a6",
)

# Load Golang rules
load("@io_bazel_rules_go//go:def.bzl", "go_repository")
go_repository()

# Docker Rules

git_repository(
    name = "io_bazel_rules_docker",
    remote = "https://github.com/bazelbuild/rules_docker.git",
    tag = "v0.7.0",
)

# Load Golang docker rules
load(
    "@io_bazel_rules_docker//docker:docker.bzl",
    "docker_repositories", "docker_pull"
)
docker_repositories()
docker_pull(
    name = "alpine",
    registry = "index.docker.io",
    repository = "library/alpine",
    digest = "sha256:769fddc7cc2f0a1c35abb2f91432e8beecf83916c421420e6a6da9f8975464b6"
)

# download the gazelle tool
http_archive(
    name = "bazel_gazelle",
    url = "https://github.com/bazelbuild/bazel-gazelle/releases/download/0.11.0/bazel-gazelle-0.11.0.tar.gz",
    sha256 = "92a3c59734dad2ef85dc731dbcb2bc23c4568cded79d4b87ebccd787eb89e8d0",
)

# load gazelle - Bazel BUILD file generator
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")
gazelle_dependencies()


# external dependencies

go_repository(
    name = "com_github_golang_protobuf",
    importpath = "github.com/golang/protobuf",
    tag = "v1.3.1",
)
```

- Scaffolding Dependencies with gazelle
```
cat <<EOF > BUILD.bazel
load("@bazel_gazelle//:def.bzl", "gazelle")

gazelle(
    name = "gazelle",
    # you project name here!
    prefix = "github.com/brendanjryan/groupcache-bazel",
)
EOF

bazel run //:gazelle
```

