# Kube Remediator

![Kube Remediator Logo ](logo/logo.png)


## List
- [Reschedules Pods in CrashLoopBackOff](#crashloopbackoff-rescheduler)
- [Deletes unbound PVCs](#unbound-persistentvolumeclaim-cleaner)

### [CrashLoopBackOff Rescheduler](pkg/remediator/crash_loop_back_off_rescheduler.go)

Reschedules Pods in CrashLoopBackOff
- Runs every 1m (`interval` config)
- Looks for containers in CrashLoopBackOff with `restartCount` > 5 (`failureThreshold` config)
- Ignores Pods without annotation `kube-remediator/CrashLoopBackOffRemediator` (`annotation` config, use `""` to manage all pods )
- Can work in a single namespace, default is all namespaces `""` (`namespace` config)
- Ignores Pods without `ownerReferences` (Avoid deleting something which does not come back)

Why:
- node issues
- stale init-container/sidecar


### Unbound PersistentVolumeClaim cleaner

Deletes left behind PersistentVolumeClaims
- Waits for 7 days(configurable) before deleting
- Ignores if `PersistentVolume` has `persistentVolumeReclaimPolicy` set to `Retain`


Why:
- When Statefulset gets deleted, associated PersistentVolumeClaims are not automatically deleted

## Configuration
Default configuration is provided under `config/*`.

## Development

Running locally on currently selected kubernetes cluster with go ~> 1.12.9:
```bash
unset GOPATH
go build -o .build/remediator cmd/remediator/app.go
.build/remediator -kubeconfig ~/.kube/config

# test remediators
kubectl apply -f kubernetes/crashloop_pod.yml
```


## Deploy

```bash
kubectl apply -f kubernetes/rbac.yaml
kubectl apply -f kubernetes/app-server.yml
```


