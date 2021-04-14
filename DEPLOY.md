# Deployment 🚀

See [Compute Default DEPLOY.md](https://github.com/zendesk/compute-defaults/blob/main/DEPLOY.md) for an overview of Compute Deployment.

| Resource  | Link  |
|:----------|:------|
| Cerebro   | <https://cerebro.zende.sk/projects/kube_remediator> |
| SLOs      | <https://zendesk.datadoghq.com/dashboard/aqi-m5k-y32/kuberemediator-slos> |

## Recovery

Revert to old stable version via [Samson](https://samson.zende.sk/projects/kube_remediator).

## Verification

Need to check the CNI Plugin has started successfully and attached an ENI to the node. 
- Check Datadog:
    * [Monitors](https://zendesk.datadoghq.com/monitors/manage?q=service%3Akube_remediator)
    * [Dashboard](https://zendesk.datadoghq.com/dashboard/m7n-a5f-jh9/kuberemediator)
- Logs:
    * [Datadog Logs](https://zendesk.datadoghq.com/logs?query=service%3Akube-remediator)
    * [Datadog Logs EU](https://app.datadoghq.eu/logs?query=service%3Akube-remediator)
    * Stern CLI: `stern -l project=kube-remediator -n kube-system`