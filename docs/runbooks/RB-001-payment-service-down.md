# Runbook RB-001: Payment Service Down

## Metadata
| Field | Value |
|-------|-------|
| **Severity** | P1 - Critical |
| **Service** | Payment Service |
| **Runbook Owner** | Platform SRE Team |
| **Last Updated** | 2024-01-15 |
| **Escalation** | #sre-critical Slack → PagerDuty on-call → Engineering Manager |

## Symptoms
- `payment-service` pods showing `CrashLoopBackOff` or `NotReady`
- Alert: `payment_service_availability < 99.99%`
- Alert: `payment_latency_p99 > 150ms` for > 2 minutes
- Customer complaints: "MADA transactions failing at checkout"

## Impact
- **Revenue**: Direct payment failure = lost sales
- **Compliance**: SAMA requires payment system availability logs
- **Downstream**: Order Service saga compensating transactions will cascade-cancel orders

## Diagnostic Steps

### 1. Check Pod Status (1 minute)
```bash
kubectl get pods -n production -l app=payment-service
kubectl describe pod <failing-pod> -n production
kubectl logs -n production -l app=payment-service --tail=200 | grep ERROR
```

### 2. Check Kafka Consumer Lag (1 minute)
```bash
kubectl exec -it kafka-client -- kafka-consumer-groups.sh \
  --bootstrap-server kafka:9092 \
  --describe --group payment-service
```
- If lag > 1000: Consumer is slow or dead

### 3. Check Database Connection Pool (1 minute)
```bash
kubectl exec -it <payment-pod> -n production -- curl localhost:8080/actuator/metrics/jdbc.connections.active
```
- If active == max: Connection pool exhaustion → check for slow queries

### 4. Check External Provider Status (2 minutes)
- MADA Status Page: https://status.mada.com.sa (verify)
- STC Pay API health: `curl https://api.stcpay.com.sa/health`

## Remediation

### Scenario A: Pod Crash (OOMKilled)
```bash
# Check memory usage
kubectl top pod -n production -l app=payment-service

# If memory limit reached:
kubectl patch deployment payment-service -n production -p \
  '{"spec":{"template":{"spec":{"containers":[{"name":"payment-service","resources":{"limits":{"memory":"2Gi"}}}]}}}}'

# Root cause: Check for memory leak in recent deployment
```

### Scenario B: Database Connection Pool Exhaustion
```bash
# Check for long-running queries
kubectl exec -it postgres-payment-0 -- psql -U payment_user -c \
  "SELECT pid, state, query_start, query FROM pg_stat_activity WHERE state = 'active' ORDER BY query_start;"

# Kill blocking queries if safe (NOT payment transactions!)
# Scale connection pool temporarily:
kubectl set env deployment/payment-service -n production SPRING_DATASOURCE_HIKARI_MAXIMUM_POOL_SIZE=50
```

### Scenario C: Kafka Consumer Rebalance Storm
```bash
# Check for frequent rebalances
kubectl logs -n production -l app=payment-service | grep "rebalance"

# Mitigation: Increase session timeout temporarily
kubectl set env deployment/payment-service -n production KAFKA_SESSION_TIMEOUT_MS=45000
```

### Scenario D: External Provider Outage (MADA/STC Pay)
```bash
# Enable circuit breaker fallback
# All new payments automatically queued for retry
kubectl set env deployment/payment-service -n production PAYMENT_CIRCUIT_BREAKER_ENABLED=true

# Notify business operations
slack-post #payments-ops "MADA outage detected. Payments queued for retry. ETA: TBD"
```

## Verification
After remediation, verify:
```bash
# Health checks passing
curl https://api.fabric.sa/health

# SLO dashboard green
open https://grafana.fabric.sa/d/payment-slo

# Kafka lag < 100
kubectl exec -it kafka-client -- kafka-consumer-groups.sh \
  --bootstrap-server kafka:9092 --describe --group payment-service

# No ERROR logs for 5 minutes
stern payment-service -n production --since 5m | grep ERROR
```

## Post-Incident Actions
1. Write post-mortem in `docs/postmortems/YYYY-MM-DD-payment-outage.md`
2. Update this runbook if new failure mode discovered
3. Add chaos experiment for the failure scenario to `scripts/chaos/`
4. Review and adjust alert threshold if it was a false positive

## Related Links
- [ADR 003: Outbox Pattern](../adr/003-why-outbox-pattern.md)
- [SAMA Control 3.1.1: Business Continuity](../samacompliance/control-mapping.md)
- [Payment Service Architecture Diagram](../../README.md)
