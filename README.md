# Kube Remediator
Missing icon

## List
- [Reschedules Pods in CrashLoopBackOff](#crashloopbackoff-rescheduler)
- [Deletes unbound PVCs](#unbound-persistentvolumeclaim-cleaner)

### CrashLoopBackOff rescheduler

Reschedules Pods in CrashLoopBackOff.
- Looks for containers in CrashLoopBackOff with `restartCount` > 5 (configurable).
- Ignores Pods without annotation `kube_remediator/CrashLoopBackOffRemediator`
- Ignores Pods without `ownerReferences` (Avoid deleting something which does not come back. TODO: recreate instead of ignoring).

Why:
- node issues.
- stale init-container/sidecar. 


### Unbound PersistentVolumeClaim cleaner

Deletes left behind PersistentVolumeClaims.
- Waits for 7 days(configurable) before deleting.
- Ignores if `PersistentVolume` has `persistentVolumeReclaimPolicy` set to `Retain`.


Why:
- When Statefulset gets deleted, associated PersistentVolumeClaims are not automatically deleted.


## Development

Running locally:
```bash
go build -o .build/remediator cmd/remediator/app.go
.build/remediator -kubeconfig ~/.kube/config 
```


## Deploy
```bash
kubectl apply -f kubernetes/rbac.yaml
kubectl apply -f kubernetes/app-server.yml
```


