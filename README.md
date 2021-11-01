This project aims to help to release a project which especially has multiple git repositories.

## Features

* Support to create a tag for a git repository
* Support to create a release (or pre-release) for [a GitHub repository](docs/github.md)
* Support to integrate with GitOps framework (such as [Argo CD](https://github.com/argoproj/argo-cd))

## Installation

Install it to a Kubernetes cluster. You can use [kubekey](https://github.com/kubesphere/kubekey) or [ks CLI](https://github.com/kubesphere-sigs/ks).

### For local environment

```shell
make deploy
```

### For production environment

TBD

## How to use

Create a secret for your git repositories with name `test-git`, such as:
```yaml
apiVersion: v1
stringData:
  password: admin
  username: admin
kind: Secret
metadata:
  name: test-git
  namespace: default
type: "kubernetes.io/basic-auth"
```

Create a Kubernetes custom resource with the following example:
```yaml
apiVersion: devops.kubesphere.io/v1alpha1
kind: Releaser
metadata:
  name: releaser-sample
spec:
  repositories:
    - name: test
      address: https://gitee.com/linuxsuren/test
      branch: master
  secret:
    name: test-git
    namespace: default
```

### Integration with ArgoCD

Please provide the corresponding git repository if you want to use GitOps way.
```yaml
apiVersion: devops.kubesphere.io/v1alpha1
kind: Releaser
metadata:
  name: releaser-sample
spec:
  gitOps:
  enable: true
  repository:
    address: https://gitee.com/linuxsuren/linuxsuren-releaser
    branch: master
    name: test
  repositories:
    - name: test
      address: https://gitee.com/linuxsuren/test
      branch: master
  secret:
    name: test-git
    namespace: default
```

Wait for a while, you can check your git repositories to see if there is a new git tag over there.
