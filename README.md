# Kube Remediator [![Build Status](https://travis-ci.com/aksgithub/kube_remediator.svg)](https://travis-ci.com/aksgithub/kube_remediator) [![coverage](https://img.shields.io/badge/coverage-100%25-success.svg)](https://github.com/aksgithub/kube_remediator)


## Remediators
- [Reschedules Pods in CrashLoopBackOff](#crashloopbackoff-rescheduler)
- [Deletes unbound PVCs](#unbound-persistentvolumeclaim-cleaner)
- [Deletes Failed Pods in Out of CPU/Memory](#failedpods-rescheduler)


### [CrashLoopBackOff Rescheduler](pkg/remediator/crashloopbackoffrescheduler.go)

Reschedules `CrashLoopBackOff` `Pod` to fix permanent crashes caused by stale init-container/sidecar/configmap 

- Listens to Pod update events and does a Pod list
- Looks for containers in CrashLoopBackOff with `restartCount` > 5 (`failureThreshold` config)
- Ignores Pods with annotation `kube-remediator/CrashLoopBackOffRemediator: "false"`
- Can work in a single namespace, default is all namespaces `""` (`namespace` config)
- Ignores Pods without `ownerReferences` (Avoid deleting something which does not come back)


### [Old Pod Deleter](pkg/remediator/oldpoddeleter.go)

Deletes `Pods` with label `kube-remediator/OldPodDeleter=true` older than 24h


### [Failed Pods Rescheduler](pkg/remediator/failedpodsrescheduler.go)

Reschedules `Failed` `Pods` by deleting them, since they are not automatically cleaned up.

- Listens to Pod update events and does a Pod list
- Finds pods in Failed status with reason `OutOfCpu`, `OutofMemory`.
- Ignores Pods without `ownerReferences` (Avoid deleting something which does not come back)
- Ignores Pods for Jobs because they can be automatically cleaned up.
- Deletes the pods in failed status after 5 mins to have time to debug

### [Completed Pods Deleter](pkg/remediator/completedpoddeleter.go)

Deletes `Pods` that in `Completed` status for more than 24h.

### Unbound PersistentVolumeClaim cleaner TODO

Deletes `PersistentVolumeClaim` left behind by deleted `StatefulSet`, that are not automatically cleaned up otherwise

- Waits for 7 days(configurable) before deleting
- Ignores if `PersistentVolume` has `persistentVolumeReclaimPolicy` set to `Retain`


## Deploy

```bash
kubectl apply -f kubernetes/rbac.yaml
kubectl apply -f kubernetes/app-server.yml
```

Configuration options:
- Deploy provided image to use defaults under `config/*`
- Make a new image `FROM` the provided image and add/remove `config/*`
- Overwrite `config/*` with a mounted `ConfigMap`


## Development

### Boot Option A:

Run in local kubernetes with docker-for-mac

```bash
rake server
```

### Boot Option B:

Run against local kubernetes cluster with go:

```bash
unset GOPATH
go mod vendor # install into local directory instead of global path
make dev # run on cluster from $KUBECONFIG (defaults to ~/.kube/config)
```

### Test

- Run unit tests: `make test`
- Run a single suite: `go test -run TestSuiteFailedPodRescheduler github.com/aksgithub/kube_remediator/pkg/remediator`
- Run a single test: comment out all other test in the suite and run the suite. TODO: improve.

```bash
# CrashLoopBackOffRemediator: pod is rescheduled after restarting 5 times ?
kubectl apply -f examples/crashloop_pod.yml

# OldPodDeleter: pod is deleted when it gets 24h old ? (best change the 24h in the code to 1min)
kubectl apply -f examples/old_pod.yml
```

Note: failed expectation in one test can lead to other tests failing. Only run one test when debugging.
