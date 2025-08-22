# DownscalePolicy CRD

The DownscalePolicy is a custom resource that allows you to manage downscaling policies with built-in support for downtime and uptime parameters.

## Installation

First, apply the CRD to your cluster:

```bash
kubectl apply -f deployments/crds/downscalepolicy-crd.yaml
```

## Usage

Create a DownscalePolicy resource to define your scaling behavior:

```yaml
apiVersion: downscaler.io/v1
kind: DownscalePolicy
metadata:
  name: my-policy
  namespace: default
spec:
  downtime: "Mon-Fri 19:00-08:00 UTC"
  uptime: "Mon-Fri 08:00-19:00 UTC"
  downscaleReplicas: 0
  gracePeriod: "15m"
```

## Spec Fields

- `downtime`: Timespans where workloads will be scaled down (e.g. 'Mon-Fri 19:00-08:00 UTC')
- `uptime`: Timespans where workloads will be scaled up (e.g. 'Mon-Fri 08:00-19:00 UTC')
- `forceDowntime`: Timespans where workloads will be forced to scale down
- `forceUptime`: Timespans where workloads will be forced to scale up
- `downscaleReplicas`: Number of replicas to scale down to (default: 0)
- `gracePeriod`: Duration a workload must exist before being scaled (e.g. '15m')
- `exclude`: Timespans when workload should be excluded from scaling
- `excludeUntil`: Exclude workload from scaling until this timestamp

## Status

The DownscalePolicy will update its status with:

- `phase`: Current phase ("Active", "Downscaled", "Excluded")
- `currentReplicas`: Current number of replicas
- `lastScaled`: Last time the workload was scaled

## Configuration

To enable DownscalePolicy management in the GoKubeDownscaler, add it to your included resources:

```yaml
includedResources:
  - deployments
  - downscalepolicies
```

The downscaler will then periodically scan for DownscalePolicy resources and manage their scaling according to the defined parameters.