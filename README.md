# Kube Remediator

![Kube Remediator Logo ](logo/logo.png)


## Remediators
- [Reschedules Pods in CrashLoopBackOff](#crashloopbackoff-rescheduler)
- [Deletes unbound PVCs](#unbound-persistentvolumeclaim-cleaner)


### [CrashLoopBackOff Rescheduler](pkg/remediator/crash_loop_back_off_rescheduler.go)

Reschedules `CrashLoopBackOff` `Pod` to fix permanent crashes caused by stale init-container/sidecar/configmap 

- Listens to Pod update events
- Looks for containers in CrashLoopBackOff with `restartCount` > 5 (`failureThreshold` config)
- Ignores Pods with annotation `kube-remediator/CrashLoopBackOffRemediator: "false"`
- Can work in a single namespace, default is all namespaces `""` (`namespace` config)
- Ignores Pods without `ownerReferences` (Avoid deleting something which does not come back)


### Unbound PersistentVolumeClaim cleaner

Deletes `PersistentVolumeClaim` left behind by deleted `StatefulSet`, that are not automatically cleanup up otherwise

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

Running locally on currently selected kubernetes cluster with go ~> 1.12.9:
```bash
unset GOPATH
go mod vendor # install into local directory instead of global path
make build
.build/remediator # run on cluster from $KUBECONFIG (defaults to ~/.kube/config) 

# test CrashLoopBackOffRemediator by seeing if this pod is rescheduled when it crashloops after it's restarted failureThreshold times
kubectl apply -f kubernetes/crashloop_pod.yml
```
