# Runbook RB-002: Kafka Consumer Lag Spike

## Metadata
| Field | Value |
|-------|-------|
| **Severity** | P2 - High |
| **Service** | Kafka / MSK |
| **Runbook Owner** | Platform SRE Team |
| **Last Updated** | 2024-01-15 |

## Symptoms
- Alert: `kafka_consumer_lag > 10000` for any consumer group
- Alert: `kafka_consumer_group_lag_increasing` for > 5 minutes
- Downstream services processing events with significant delay

## Impact
- **Order Service**: Orders stuck in PENDING state
- **Settlement Service**: Reconciliation delayed (SAMA violation if > 1 hour)
- **Fraud Service**: Real-time scoring ineffective

## Diagnostic Steps

### 1. Identify Affected Consumer Group
```bash
kafka-consumer-groups.sh --bootstrap-server $MSK_BROKERS \
  --describe --all-groups | awk '$6 > 10000 {print}'
```

### 2. Check Consumer Group Members
```bash
kafka-consumer-groups.sh --bootstrap-server $MSK_BROKERS \
  --describe --group <group-name>
```
- If `CONSUMER-ID` column shows fewer members than partitions: Consumer crashed
- If `CURRENT-OFFSET` not advancing: Consumer stuck (deadlock/gc pause)

### 3. Check Topic Partition Distribution
```bash
kafka-topics.sh --bootstrap-server $MSK_BROKERS \
  --describe --topic <topic-name>
```
- Look for hot partitions (uneven leader distribution)

### 4. Check Consumer Service Metrics
```bash
# CPU throttling
kubectl top pod -n production -l app=<consumer-service>

# GC pauses (Java services)
kubectl logs -n production -l app=<consumer-service> | grep "GC pause"
```

## Remediation

### Scenario A: Consumer Pod Crash/Cycle
```bash
# Check pod status
kubectl get pods -n production -l app=<consumer-service>

# If CrashLoopBackOff: Follow RB-001 for that service
# If pod count < desired: Check HPA / Karpenter
kubectl get hpa -n production <consumer-service>
```

### Scenario B: Slow Message Processing
```bash
# Temporarily scale up consumers
kubectl scale deployment <consumer-service> -n production --replicas=10

# If partition count < consumer count: Consumers will be idle
# Add partitions (careful - changes partition assignment!)
kafka-topics.sh --bootstrap-server $MSK_BROKERS \
  --alter --topic <topic-name> --partitions 12
```

### Scenario C: Poison Pill Message
```bash
# Find the stuck offset
kafka-consumer-groups.sh --bootstrap-server $MSK_BROKERS \
  --describe --group <group-name>

# Inspect the message at that offset
kafka-console-consumer.sh --bootstrap-server $MSK_BROKERS \
  --topic <topic-name> --partition <partition> --offset <offset> --max-messages 1

# Skip the poison pill (LAST RESORT - log for audit)
kafka-consumer-groups.sh --bootstrap-server $MSK_BROKERS \
  --group <group-name> --topic <topic-name>:<partition> --reset-offsets --to-offset <offset+1> --execute

# Alert audit team
slack-post #audit "Poison pill message skipped. Topic=<topic> Partition=<partition> Offset=<offset>"
```

### Scenario D: Hot Partition
```bash
# Check partition leader distribution
kafka-topics.sh --bootstrap-server $MSK_BROKERS --describe --topic <topic>

# If all leaders on same broker: Reassign partitions
kafka-reassign-partitions.sh --bootstrap-server $MSK_BROKERS \
  --topics-to-move-json-file topics.json --broker-list "1,2,3" --generate
```

## Verification
```bash
# Lag should decrease within 10 minutes
kafka-consumer-groups.sh --bootstrap-server $MSK_BROKERS \
  --describe --group <group-name>

# Settlement lag SLO: < 1 hour
# Check Grafana dashboard: https://grafana.fabric.sa/d/kafka-lag
```

## Prevention
- Set up proactive alerts: `kafka_consumer_lag > 5000` (warning)
- Ensure partition count > max replica count for consumer
- Review message size: > 1MB messages cause processing delays
- Monitor consumer GC patterns during load tests
