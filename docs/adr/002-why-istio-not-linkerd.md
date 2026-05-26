# ADR 002: Why Istio instead of Linkerd for the Service Mesh

## Status
Accepted

## Context
We need a service mesh to provide:
- mTLS encryption between all 9 microservices
- L7 authorization policies (only Order Service can call Payment Service)
- Traffic mirroring for realistic staging tests
- Multi-cluster mesh across Riyadh, Dubai, and Bahrain

## Decision
We chose **Istio** over Linkerd.

## Rationale

| Criteria | Istio | Linkerd |
|----------|-------|---------|
| Authorization Policies | Rich (JWT, mTLS, RBAC) | Basic | 
| Traffic Mirroring | Native | Experimental |
| Multi-Cluster Mesh | Mature (single control plane) | Complex |
| Wasm Extensibility | Supported | Not supported |
| Resource Overhead | Higher (~100MB/sidecar) | Lower (~20MB/sidecar) |
| Enterprise Adoption | Wide (Google, IBM) | Growing (Buoyant) |

### Key Deciding Factors

1. **Authorization Policies**: SAMA requires strict service-to-service access controls. Istio's `AuthorizationPolicy` resource allows us to enforce rules like:
   ```yaml
   - from:
     - source:
         principals: ["cluster.local/ns/production/sa/order-service"]
     to:
     - operation:
         methods: ["POST"]
         paths: ["/api/v1/payments/*"]
   ```

2. **Multi-Cluster Mesh**: We need a single logical mesh across Riyadh (primary), Dubai (DR), and Bahrain (DR). Istio's multi-cluster single control plane model is production-proven at this scale.

3. **Traffic Mirroring**: Critical for our chaos engineering practice. We mirror 1% of production traffic to staging to validate releases with real data patterns.

## Consequences

### Positive
- Zero-trust network with automatic mTLS certificate rotation
- Rich telemetry for SRE dashboards
- Canary analysis with Flagger leveraging Istio traffic splitting

### Negative
- Higher memory footprint per pod (~100MB sidecar)
- Steeper learning curve for platform engineers
- Control plane complexity in multi-cluster topology

## Mitigations
- Karpenter auto-scaling compensates for sidecar overhead
- Dedicated Istio runbook and training for on-call engineers
- Control plane deployed in Riyadh with documented DR promotion procedure
