# ADR 004: Why Karpenter instead of Cluster Autoscaler

## Status
Accepted

## Context
Saudi e-commerce experiences extreme traffic spikes during:
- Ramadan (pre-iftar ordering, ~3x normal traffic)
- White Friday (single-day 10x spike)
- Government salary days (predictable monthly 2x spikes)

We need node provisioning in seconds, not minutes, to handle these flash sales without over-provisioning.

## Decision
We chose **Karpenter** over Kubernetes Cluster Autoscaler.

## Comparison

| Criteria | Karpenter | Cluster Autoscaler |
|----------|-----------|-------------------|
| Node Provisioning Speed | 15-30 seconds | 2-5 minutes |
| Instance Type Flexibility | Native (multi-arch, spot, Graviton) | Requires node pools per type |
| Bin Packing Efficiency | Direct pod-to-node scheduling | Pre-defined ASG sizes |
| Consolidation | Automatic (removes underutilized nodes) | Manual or basic scale-down |
| AWS Integration | Native (launches instances directly) | Via Auto Scaling Groups |
| Maturity | v0.32+ (AWS GA) | Very mature |

## Rationale

### 1. Flash Sale Response Time
During White Friday 2023, our load tests showed:
- Cluster Autoscaler: pods stayed Pending for 3-4 minutes → customer timeouts → abandoned carts
- Karpenter: pods scheduled in 15-20 seconds → maintained p99 latency < 200ms

### 2. Cost Optimization
Karpenter's consolidation feature automatically:
- Replaces multiple small instances with fewer large instances
- Moves workloads to Spot when possible
- Uses Graviton3 (ARM64) for 20% cost savings

Our cost modeling showed **35% EC2 savings** versus Cluster Autoscaler with fixed node pools.

### 3. Simplified Operations
No need to maintain multiple ASGs (t3.medium for system, m6i.large for general, c6i.xlarge for compute). Karpenter selects the optimal instance type per pod requirements.

## Consequences

### Positive
- Faster scale-out for traffic spikes
- Lower infrastructure costs via consolidation
- Support for diverse instance types without operational overhead

### Negative
- Less mature ecosystem (fewer StackOverflow answers)
- Tighter AWS coupling (not portable to GCP/Azure)
- Requires careful resource request/limit settings (Karpenter uses these for scheduling decisions)

## Mitigations
- Extensive load testing with k6 before each major sale event
- Runbook for Karpenter-specific issues (`docs/runbooks/RB-005-karpenter-scale-failure.md`)
- Fallback node pool with Cluster Autoscaler for critical system pods if Karpenter fails

## Evidence
```bash
# Karpenter node provisioning time
kubectl logs -n karpenter deployment/karpenter | grep "launched node"
# Typical output: "launched node in 18.45s"
```
