# Introduction
This is sample operator to deploy and manage CPU arch emulator (eg. qemu)
lifecycle in a Kubernetes cluster

## Install

### Prerequisites
Ensure KUBECONFIG is pointing to a working Kubernetes cluster

### Deploy the Operator
```
$ git clone https://github.com/bpradipt/arch-emulator-operator.git
$ cd arch-emulator-operator
$ make install && make deploy IMG=quay.io/bpradipt/arch-emulator-operator
```

This will deploy the controller POD in the `arch-emulator-operator-system`
namespace

```
$ kubectl get pods -n arch-emulator-operator-system

NAME                                                        READY   STATUS    RESTARTS   AGE
arch-emulator-operator-controller-manager-cdf6df6d9-9rhg9   2/2     Running   0          19m
```

### Create a Custom Resource
```
$ kubectl create -f config/samples/emulator_v1alpha1_archemulator.yaml
```
This will create a `job` to download and run the emulator setup on the nodes

```
$kubectl get all

NAME                            READY   STATUS      RESTARTS   AGE
pod/archemulator-sample-744bk   0/1     Completed   0          8m4s

NAME                 TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
service/kubernetes   ClusterIP   10.96.0.1    <none>        443/TCP   7d4h

NAME                            COMPLETIONS   DURATION   AGE
job.batch/archemulator-sample   1/1           2s         8m4s
```

## Uninstall
```
$ make uninstall
```

## Hacking

### Build and Push the container image

Ensure you have access to a container registry like quay.io or hub.docker.com
```
export REGISTRY=<registry>
export REGISTRY_USER=<user>
```

```
$ make docker build IMG=${REGISTRY}/${REGISTRY_USER}/arch-emulator-operator
$ make docker-push IMG=${REGISTRY}/${REGISTRY_USER}/arch-emulator-operator
```

### Deploy the new image
```
$ make install && make deploy IMG=${REGISTRY}/${REGISTRY_USER}/arch-emulator-operator
```


