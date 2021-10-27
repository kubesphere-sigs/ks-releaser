This project aims to help to release a project which especially has multiple git repositories.

## Installation

Install it to a Kubernetes cluster. You can use [kubekey](https://github.com/kubesphere/kubekey) or [ks CLI](https://github.com/kubesphere-sigs/ks).

```shell
kubectl apply -f https://raw.githubusercontent.com/kubesphere-sigs/ks-releaser/master/config/crd/bases/devops.kubesphere.io_releasers.yaml
kubectl apply -f https://raw.githubusercontent.com/kubesphere-sigs/ks-releaser/master/config/samples/deployment.yaml
```

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

Wait for a while, you can check your git repositories to see if there a new git tag over there.
