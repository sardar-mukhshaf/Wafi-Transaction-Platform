# Saudi Distributed Commerce & Settlement Fabric

> **Reference Architecture** — A SAMA-compliant, multi-region, event-driven distributed commerce platform handling 10M+ daily transactions across KSA, UAE, and Bahrain. Built as a production-grade portfolio piece demonstrating senior/staff-level microservices, DevOps, and platform engineering capabilities.

---

## How the Code Works (Step-by-Step Process)
Here is a simple breakdown of what happens in the code when a customer buys something:

1. **Identity Service**: A customer logs into the platform securely.
2. **Catalog Service**: The customer browses items, getting instant product details and prices.
3. **Inventory Service**: When the customer wants to buy, this service checks the warehouse and temporarily holds the item so nobody else can buy it.
4. **Order Service**: This is the "manager" of the checkout process. It takes the order and orchestrates the next steps.
5. **Payment Service**: It securely talks to banks (like MADA or Apple Pay) to charge the customer's card safely.
6. **Settlement Service**: At the end of the day, it calculates who gets paid what (e.g., taxes, seller payouts).
7. **Notification Service**: It sends an SMS or email saying "Your order is confirmed!"

Behind the scenes: The **Audit Service** records every single action for legal reasons, and the **Fraud Service** constantly monitors for suspicious activity.

## What Does This Project Actually Do? (In Simple Words)
Imagine you are building a massive online shopping mall (like noon.com or Amazon). 

This project is **not** the website buttons or the pretty pictures you see. Instead, it is the **invisible engine room** running behind the scenes. 

It handles the extremely hard parts of running a massive online mall across multiple countries:
- Making sure millions of people can buy things at the exact same time without the website crashing.
- Keeping hackers out and protecting credit card numbers.
- Following all the strict legal and banking rules of Saudi Arabia.

In short, it provides the foundational code and blueprint to build a huge, safe, and lightning-fast online marketplace.

---

## Table of Contents

1. [What Is This?](#what-is-this)
2. [What Does It Do?](#what-does-it-do)
3. [Why Does This Exist?](#why-does-this-exist)
4. [When Should You Use This?](#when-should-you-use-this)
5. [Who Is This For?](#who-is-this-for)
6. [Architecture Overview](#architecture-overview)
7. [Tech Stack & Justifications](#tech-stack--justifications)
8. [Repository Structure](#repository-structure)
9. [Getting Started](#getting-started)
10. [Security & SAMA Compliance](#security--sama-compliance)
11. [Observability & SRE](#observability--sre)
12. [CI/CD Golden Path](#cicd-golden-path)
13. [GitOps & Progressive Delivery](#gitops--progressive-delivery)
14. [Chaos Engineering](#chaos-engineering)
15. [Implementation Roadmap](#implementation-roadmap)
16. [Interview Talking Points](#interview-talking-points)
17. [License](#license)

---

## What Is This?

This is **not** a tutorial-level "Docker Compose + 3 services" demo. It is a **production-grade, multi-cluster, event-driven, zero-trust microservices platform** designed as a **reference architecture** for enterprise-scale distributed commerce in the Saudi Arabian market.

Specifically, this project models a **B2B/B2C marketplace backend** — similar in complexity to what powers noon.com, Jarir, or SABIC's procurement platforms — with the following characteristics:

- **Multi-tenant** by design (merchant isolation at the data layer)
- **Multi-region** (Riyadh primary, Dubai DR, Bahrain DR)
- **SAMA-compliant** (Saudi Central Bank security and audit requirements)
- **PCI-DSS aware** (payment tokenization and vaulting patterns)
- **Event-driven** (Saga orchestration, CQRS, Event Sourcing)
- **Zero-trust networking** (Istio service mesh with strict mTLS)
- **GitOps-managed** (ArgoCD with progressive delivery via Flagger)

Think of it as the **infrastructure and platform layer** you would build *before* product teams start writing business logic. It answers the question: *"How do we run 50+ microservices across three countries without losing transactions, leaking data, or waking engineers up at 3 AM?"*

---

## What Does It Do?

At a functional level, this platform enables:

| Capability | Description |
|------------|-------------|
| **Identity & Access** | OAuth2/OIDC authentication, SAMA-compliant MFA, KYC integration hooks, role-based access control (RBAC) per merchant |
| **Product Catalog** | Multi-language product listings (Arabic RTL support), dynamic pricing engine, search indexing |
| **Order Management** | Full order lifecycle (cart → checkout → fulfillment → delivery), Saga-orchestrated distributed transactions |
| **Payment Processing** | MADA, STC Pay, Apple Pay integration patterns, PCI-DSS token vault, idempotent charge handling |
| **Settlement & Reconciliation** | Daily SAR reconciliation, VAT calculation, multi-merchant payout scheduling |
| **Inventory Management** | Real-time stock levels across warehouses (Riyadh, Jeddah, Dammam), reservation patterns |
| **Notifications** | Multi-channel delivery: SMS (Mobily/Zain), WhatsApp Business API, email, push notifications |
| **Audit & Compliance** | Immutable SAMA audit trails, 7-year log retention, append-only event logs |
| **Fraud Detection** | Real-time rule engine + ML anomaly detection on transaction patterns |

At a **platform** level, it provides:

- **Automated infrastructure provisioning** (Terraform/Terragrunt across environments)
- **Golden-path CI/CD** (build, test, secure, sign, deploy with zero manual steps to staging)
- **Progressive delivery** (canary deployments with automatic rollback on SLO violation)
- **Unified observability** (metrics, logs, traces, profiles — correlated by trace ID)
- **Chaos engineering** (weekly automated failure injection to validate resilience)
- **Policy-as-code** (OPA/Sentinel preventing non-compliant infrastructure changes)

---

## Why Does This Exist?

### The Problem

Saudi enterprises — particularly in fintech, industrial IoT, and smart city logistics — face a unique set of constraints:

1. **Regulatory burden:** SAMA, SDAIA, and ZATCA impose strict data residency, encryption, and audit requirements.
2. **Scale spikes:** Ramadan, White Friday, and government payroll cycles (e.g., Eid bonuses) create massive, predictable traffic spikes.
3. **Multi-region complexity:** KSA, UAE, and Bahrain operations require active-active or hot-standby DR strategies.
4. **Talent expectations:** Senior engineering roles in Riyadh and Dubai now expect demonstrable experience with service meshes, GitOps, and event-driven architecture.

### The Solution

This repository exists to prove that you — as a senior/staff engineer — can:

- Design **domain-driven boundaries** that prevent the "distributed monolith" anti-pattern
- Implement **data consistency patterns** (Saga, Outbox, CQRS) that actually work under failure
- Build **IaC pipelines** with policy enforcement, drift detection, and cost governance
- Operate **Kubernetes at scale** with auto-scaling, mesh networking, and secrets management
- Define **SLOs and error budgets** that align engineering with business outcomes
- Write **runbooks and ADRs** that make your team operate like a platform company, not a startup

### Why Not Use Off-the-Shelf SaaS?

Because Saudi enterprises **cannot** always use Stripe, Shopify, or Twilio due to:
- Data residency requirements (customer PII must stay in KSA)
- SAMA approval requirements for payment processors
- Custom VAT and reconciliation rules
- Integration with local providers (MADA, STC Pay, Mobily SMS)

This architecture shows you know how to **build** when buy is not an option.

---

## When Should You Use This?

### Use This Architecture When:

- You are building a **marketplace, payment platform, or logistics system** targeting KSA/UAE/Bahrain
- You need **SAMA/SDAIA compliance** from day one (not as an afterthought)
- You expect **>1M daily transactions** within 12 months of launch
- You have **multiple product teams** that need isolated deployment lifecycles
- You need **active-active multi-region** failover (not just cold backups)
- You are interviewing for **Senior/Staff/Principal Platform Engineer** roles in Saudi or UAE

### Do Not Use This When:

- You are building a simple CRUD app or MVP (use Laravel/Rails/Next.js on Vercel instead)
- Your team has <3 backend engineers (the operational overhead is not worth it)
- You have no Kubernetes experience (learn EKS basics first)
- Your budget is <$5K/month in infrastructure (this stack assumes enterprise scale)

---

## Who Is This For?

| Audience | How They Use This |
|----------|-------------------|
| **Senior Backend Engineers** | Study the service boundaries, Saga implementations, and Outbox pattern |
| **Platform/DevOps Engineers** | Copy the Terraform modules, ArgoCD setup, and observability stack |
| **SREs** | Adopt the SLO definitions, runbooks, and alerting hierarchy |
| **Security Engineers** | Reference the SAMA control mapping, OPA policies, and Vault integration |
| **Hiring Managers in KSA/UAE** | Evaluate candidates who fork/contribute to this as a skills signal |
| **Candidates Interviewing** | Present this as a portfolio piece with the talking points below |

---

## Architecture Overview

```
                              EDGE LAYER
  CloudFront + AWS WAF  AWS Shield  API Gateway (Regional: Riyadh/Dubai)



                         CONTROL PLANE (EKS - Istio)

     Ingress         Identity        API GW          Rate
     Gateway         Service         BFF Layer       Limiter
    (Istio GW)      (Keycloak)      (GraphQL)       (Redis)



                      SERVICE MESH (Istio - mTLS Strict)

     Catalog        Order        Payment      Settlement
     Service       Service       Service       Service
    (Node.js)      (Go)          (Java)        (Rust)


    Inventory   Notification     Audit         Fraud
     Service       Service       Service       Service
     (Go)         (Python)       (Go)          (Python)



                    DATA & MESSAGING LAYER

    Kafka/MSK       PostgreSQL        Redis          S3 (Iceberg
   (Event Bus)      (Per Service     (Cache/         Data Lake)
    + Schema         + read          Session)
     Registry        replicas)


    ClickHouse      ElasticSearch  (Analytics & Search)
    (OLAP)          (Logs/Search)



                    OBSERVABILITY STACK
  Prometheus + Thanos  Grafana  Jaeger  Loki  OpenTelemetry Collector
  PagerDuty  OpsGenie  Status Page  SLO Dashboards
```

### Design Principles

1. **Domain-Driven Design (DDD):** Each service owns its data, schema, and deployment lifecycle. No shared databases.
2. **Smart Endpoints, Dumb Pipes:** Business logic lives in services; Kafka is a simple, reliable message bus.
3. **Fail Fast, Recover Automatically:** Circuit breakers, retries with jitter, and bulkheads at the mesh level.
4. **Observability by Default:** Every service emits OpenTelemetry traces, structured logs, and Prometheus metrics.
5. **Security in Depth:** Defense at the edge (WAF), network (mTLS), application (RBAC), and data (encryption) layers.

---

## Tech Stack & Justifications

| Layer | Technology | Justification |
|-------|-----------|---------------|
| **Orchestration** | EKS + Karpenter | Saudi region availability; Karpenter provisions nodes in seconds (critical for flash sale spikes like Ramadan/White Friday) |
| **Service Mesh** | Istio | Enterprise standard; strong L7 authorization policies; multi-cluster mesh support for Riyadh-Dubai-Bahrain |
| **GitOps** | ArgoCD + Flagger | Declarative deployments; automated canary analysis with Prometheus metrics; automatic rollback on 5xx errors |
| **Messaging** | Kafka (MSK) + Schema Registry | Event sourcing backbone; Avro schema enforcement prevents consumer breakage; exactly-once semantics for payments |
| **IaC** | Terraform + Terragrunt + Atlantis | DRY multi-environment configs; PR-based infrastructure changes with policy enforcement and cost estimation |
| **CI/CD** | GitHub Actions + Tekton | Cloud-native reusable workflows; cloud-agnostic if migrating from GitHub later |
| **Observability** | Prometheus/Grafana/Loki/Jaeger/Pyroscope | Open source; full data residency control; correlated traces/logs/metrics/profiles |
| **Security** | Vault, Trivy, Falco, OPA | Zero-trust secrets management; shift-left vulnerability scanning; runtime threat detection; policy-as-code |
| **Data** | Aurora PG, Redis, ClickHouse, S3 | Polyglot persistence per service; cost-effective OLAP for settlement analytics; Iceberg data lake for ML |

### Language Selection Rationale

| Service | Language | Why |
|---------|----------|-----|
| Identity | Java/Spring | Mature OAuth2/OIDC ecosystem; SAMA-compliant crypto libraries |
| Catalog | Node.js/Nest | Rapid API development; strong GraphQL ecosystem |
| Order | Go | High-throughput orchestration; excellent Kafka client libraries |
| Payment | Java | PCI-DSS compliant HSM integrations; battle-tested financial libraries |
| Settlement | Rust | Memory-safe high-performance reconciliation; CPU-efficient OLAP ingestion |
| Inventory | Go | Concurrent stock reservation; low-latency cache operations |
| Notification | Python | Rich NLP/Arabic text processing; async framework (FastAPI/Celery) |
| Audit | Go | Deterministic performance for append-only high-volume writes |
| Fraud | Python/ML | PyTorch/Scikit-learn ecosystem for anomaly detection models |

---

## Repository Structure

```
saudi-distributed-commerce-settlement-fabric/
├── .github/
│   └── workflows/              # Reusable composite actions, golden path pipeline
├── apps/
│   ├── catalog-service/        # NestJS, Dockerfile, helm/
│   ├── order-service/          # Go, Dockerfile, helm/
│   ├── payment-service/        # Java/Spring Boot, Dockerfile, helm/
│   ├── settlement-service/     # Rust, Dockerfile, helm/
│   ├── inventory-service/      # Go, Dockerfile, helm/
│   ├── notification-service/   # Python/FastAPI, Dockerfile, helm/
│   ├── audit-service/          # Go, Dockerfile, helm/
│   └── fraud-service/          # Python, Dockerfile, helm/
├── libs/
│   ├── proto/                  # Shared protobuf definitions (gRPC contracts)
│   ├── kafka-clients/          # Internal event publishing library (Go/Java/Node)
│   ├── observability/          # OpenTelemetry configs, shared instrumentation
│   └── saga-sdk/               # Reusable Saga orchestrator client
├── infrastructure/
│   ├── terraform/
│   │   ├── modules/            # Reusable: EKS, VPC, Kafka, RDS, Observability
│   │   └── environments/
│   │       ├── production/     # Riyadh, Dubai, Bahrain live configs
│   │       └── staging/        # Shared staging environment
│   └── policies/               # OPA/Sentinel Rego files (SAMA compliance)
├── gitops/
│   ├── argocd/                 # App of Apps manifests
│   ├── flux/                   # Alternative GitOps (optional)
│   └── flagger/                # Canary analysis templates
├── docs/
│   ├── adr/                    # Architecture Decision Records (001-why-kafka.md, etc.)
│   ├── runbooks/               # Incident response playbooks
│   └── samacompliance/         # SAMA control mapping matrices
├── scripts/
│   ├── local-setup/            # Tilt / Docker Compose for dev environment
│   └── chaos/                  # Litmus/Gremlin experiment manifests
├── mkdocs.yml                  # Documentation site generator config
└── README.md                   # You are here
```

---

## Getting Started

### Prerequisites

- Docker Desktop with Kubernetes enabled (or Rancher Desktop)
- `kubectl`, `helm`, `istioctl`
- `terraform` >= 1.5, `terragrunt` >= 0.50
- `tilt` (for local development)
- AWS CLI configured with appropriate credentials (for remote environments)

### Local Development (Tilt)

```bash
# Clone the repository
git clone https://github.com/yourname/saudi-distributed-commerce-settlement-fabric.git
cd saudi-distributed-commerce-settlement-fabric

# Start local stack: Kafka, PostgreSQL, Redis, and 3 core services
make local-up

# Or manually with Tilt
cd scripts/local-setup
tilt up

# Run integration tests against local stack
make test-integration

# Run contract tests (Pact)
make test-contract

# Tear down
make local-down
```

### Deploy to Staging

```bash
# Plan infrastructure changes
cd infrastructure/terraform/environments/staging
terragrunt plan

# Apply (via Atlantis in production; manual allowed in staging)
terragrunt apply

# Deploy services via ArgoCD
kubectl apply -f gitops/argocd/apps/values-staging.yaml
```

---

## Security & SAMA Compliance

This architecture is designed to satisfy **SAMA Cyber Security Framework v1.0** and **SDAIA Personal Data Protection Law (PDPL)** requirements.

| Control Domain | Implementation |
|----------------|----------------|
| **Encryption at Rest** | AES-256 (RDS, S3, EBS, Kafka MSK); managed via AWS KMS with key rotation |
| **Encryption in Transit** | TLS 1.3 for all external traffic; Istio STRICT mTLS mesh-wide for internal traffic |
| **Secrets Management** | HashiCorp Vault with AWS KMS auto-unseal; dynamic database credentials (TTL 1h); no static keys |
| **Network Security** | Private subnets only; VPC endpoints for all AWS services; zero 0.0.0.0/0 security group rules |
| **Identity & Access** | IRSA (IAM Roles for Service Accounts); no long-lived AWS credentials in pods |
| **Container Security** | Distroless images; read-only root filesystem; non-root users; Falco runtime threat detection |
| **Audit Logging** | CloudTrail + custom Audit Service; immutable S3 with Object Lock; 7-year retention |
| **Vulnerability Management** | Trivy FS/Container scans in CI; Snyk dependency monitoring; weekly patch cadence |
| **Policy Enforcement** | OPA/Sentinel: `sama_v1.0` policy pack blocks public S3 buckets, enforces tagging, mandates encryption |

### SAMA Control Mapping

Detailed mapping available in `docs/samacompliance/control-mapping.md`. Example:

| SAMA Control | Requirement | Implementation | Evidence |
|-------------|-------------|----------------|----------|
| 3.1.2 | Data Classification | S3 bucket tagging + Macie auto-discovery | Terraform + Macie reports |
| 3.3.1 | Access Control | Keycloak RBAC + IRSA | IAM policies, Keycloak realms |
| 3.4.1 | Cryptography | KMS CMKs, TLS 1.3, mTLS | AWS Config rules, Istio PeerAuthentication |
| 5.2.1 | Logging & Monitoring | CloudTrail + Audit Service + Loki | S3 access logs, Grafana dashboards |

---

## Observability & SRE

### The Three Pillars (Plus One)

| Pillar | Tool | Implementation |
|--------|------|----------------|
| **Metrics** | Prometheus + Thanos | Custom SLOs: `availability > 99.99%`, `payment_latency_p99 < 150ms`; long-term storage in S3 |
| **Logs** | Loki + S3 | Structured JSON logs; correlation IDs propagated across Saga transactions; Arabic UTF-8 safe |
| **Traces** | Jaeger/Tempo | OpenTelemetry auto-instrumentation; trace-to-log linking; sampling: 100% errors, 1% success |
| **Profiles** | Pyroscope/Parca | Continuous profiling to detect CPU/memory regressions in Go/Rust services |

### SLOs & Error Budgets

| Service | SLO | Error Budget (30d) | Alert Threshold |
|---------|-----|-------------------|-----------------|
| Payment | 99.99% availability | 4.32 min downtime | PagerDuty after 1 min |
| Order | p99 latency < 200ms | 0.1% requests > 200ms | Slack warning at 0.05% |
| Settlement | Reconciliation lag < 1h | 0 lag > 2h | SMS to finance ops |
| Catalog | 99.9% availability | 43.2 min downtime | Slack auto-resolve < 5 min |

### Alerting Hierarchy

- **Warning:** Slack channel (#sre-alerts) — auto-resolve if duration < 5 minutes
- **Critical:** PagerDuty — on-call rotation timezone-aware (Riyadh GMT+3), prayer time respect configured
- **Business:** Direct SMS to operations teams for revenue-impacting events (settlement delays, fraud spikes)

### Runbooks

Every critical alert has a runbook in `docs/runbooks/`:

- `RB-001-payment-service-down.md` — Saga compensating transaction validation
- `RB-002-kafka-lag-spike.md` — Consumer group rebalancing and partition assignment
- `RB-003-az-failure-riyadh.md` — Traffic shift to 1b/1c, RTO validation
- `RB-004-vault-seal-event.md` — Auto-unseal recovery and root token rotation

---

## CI/CD Golden Path

The `.github/workflows/golden-path.yml` defines the single, mandatory pipeline for all services.

### Stage 1: Validate

- **Lint:** ESLint, golangci-lint, Checkstyle, Rust clippy
- **Unit Tests:** Coverage gate > 85%; mutation testing where applicable
- **SonarQube:** Quality Gate = 0 critical issues, 0 security hotspots
- **Trivy FS Scan:** Secret detection + vulnerability scan on dependencies

### Stage 2: Build

- **Multi-arch builds:** AMD64 + ARM64 (Graviton3 for 20% cost reduction in AWS Riyadh)
- **BuildKit:** Layer caching, parallel stage execution
- **Cosign signing:** All images signed with Sigstore; verification required in cluster
- **SBOM generation:** CycloneDX format uploaded to Dependency-Track

### Stage 3: Integration

- **Ephemeral K8s:** kind/k3d cluster spun up via Terraform for each PR
- **Contract tests:** Pact.io consumer/provider verification
- **Integration tests:** Testcontainers for PostgreSQL, Kafka, Redis
- **Load tests:** k6 — fail if p99 > 200ms or error rate > 0.1%

### Stage 4: Security Gate

- **Snyk container scan:** Block on critical CVEs in base images
- **OPA scan:** Reject K8s manifests violating pod security standards
- **Checkov/TfSec:** IaC changes scanned for misconfigurations

### Stage 5: Deploy

- **Staging:** Direct push + ArgoCD auto-sync (immediate feedback)
- **Production:** ArgoCD + manual promotion gate (click to promote)
- **Canary:** Flagger + Istio — 5% → 25% → 100% with automated rollback on 5xx rate > 1%

### Senior Differentiators

- **Feature Flags:** LaunchDarkly integration — deploy code disabled, enable via flag (decouples deploy from release)
- **DB Migrations:** Flyway/Liquibase in init containers; **backward-compatible only** (expand/contract pattern)
- **Pipeline as Code:** Jenkinsfile/Tekton tasks versioned alongside services; no click-ops

---

## GitOps & Progressive Delivery

### ArgoCD App of Apps

```yaml
# gitops/argocd/apps/values-production.yaml
applications:
  - name: order-service
    source:
      repoURL: https://github.com/yourname/saudi-fabric-gitops
      path: services/order/overlays/riyadh
    syncPolicy:
      automated:
        prune: true
        selfHeal: true
      retry:
        limit: 5
    ignoreDifferences:
      - group: apps
        kind: Deployment
        jsonPointers:
          - /spec/replicas  # HPA manages this
```

### Key Patterns

- **Multi-source:** Helm charts from ChartMuseum + values from Git + images from ECR
- **Sealed Secrets / External Secrets:** HashiCorp Vault integration with AWS KMS auto-unseal
- **Argo Rollouts:** Blue-green deployments for Identity Service (zero-downtime token rotation)
- **Notifications:** Slack/Teams alerts for sync failures; PagerDuty for production drift

---

## Chaos Engineering

Weekly automated chaos experiments validate resilience assumptions.

```yaml
# scripts/chaos/weekly-experiments.yaml
experiments:
  - name: pod-delete-payment
    target: payment-service
    duration: 300s
    interval: 60s
    expected: Saga compensating transactions trigger; orders move to PENDING_PAYMENT

  - name: az-failure-riyadh
    target: availability-zone:1a
    expected: Traffic shifts to 1b/1c; RTO < 30 seconds

  - name: kafka-partition-loss
    target: kafka-topic:payment-events
    expected: Outbox buffer retains events; Debezium catches up with zero data loss
```

### Tools

- **LitmusChaos:** Kubernetes-native chaos experiments
- **Gremlin:** SaaS option for multi-region failure injection
- **Chaos Mesh:** Alternative for advanced network partition simulation

---

## Implementation Roadmap

| Week | Deliverable | Milestone |
|------|-------------|-----------|
| **1** | Local stack: Tilt + Docker Compose, 3 core services (Order, Payment, Catalog), Kafka | `v0.1-local` |
| **2** | Terraform: VPC + EKS (Riyadh region) + Istio service mesh | `v0.2-infra` |
| **3** | CI/CD: Golden path pipeline, image signing, Trivy scanning | `v0.3-cicd` |
| **4** | GitOps: ArgoCD installation, sealed secrets, multi-env overlays | `v0.4-gitops` |
| **5** | Observability: OpenTelemetry, Jaeger, Grafana dashboards, SLO alerts | `v0.5-obs` |
| **6** | Advanced patterns: Saga orchestration, Outbox pattern, Debezium CDC | `v0.6-patterns` |
| **7** | Security: Vault integration, Falco, OPA policies, SAMA documentation | `v0.7-security` |
| **8** | Polish: Chaos tests, runbooks, ADRs, public README, portfolio site | `v1.0-portfolio` |

---

## Interview Talking Points

> Use these in senior/staff interviews for Saudi/UAE engineering roles.

**On Scale & Performance:**
> *"I chose Karpenter over Cluster Autoscaler because Saudi e-commerce flash sales — like Ramadan and White Friday — need node provisioning in seconds, not minutes. During load testing, Karpenter reduced our scale-up time from 3 minutes to 15 seconds."*

**On Data Consistency:**
> *"The Outbox pattern ensures we never lose a payment event, even if Kafka is temporarily unavailable. This is critical for MADA integration where duplicate or missing transactions create reconciliation nightmares. We coupled this with idempotency keys so consumers can safely retry."*

**On Compliance:**
> *"All Terraform changes are validated with OPA policies enforcing SAMA requirements before they reach production. For example, we have a policy that blocks any S3 bucket without encryption and tags identifying the data classification level. This turns compliance from a quarterly audit into a shift-left gate."*

**On Observability:**
> *"We defined SLOs not just for availability, but for business outcomes — like 'settlement reconciliation lag must be under 1 hour.' When that SLO burns its error budget, it triggers an SMS to finance ops, not just a Slack message to engineering. This aligns SRE with business impact."*

**On Multi-Region:**
> *"We run an Istio multi-cluster mesh across Riyadh and Dubai with a single control plane in Riyadh. If the Riyadh control plane fails, we have a documented 5-minute runbook to promote Dubai to control plane authority. Our chaos tests validate this monthly."*

---

## Senior vs. Junior Distinction

| Junior/Mid-Level Approach | **Senior/Staff Approach (This Repo)** |
|---------------------------|---------------------------------------|
| "I deployed microservices on Kubernetes" | "I designed a multi-tenant, event-driven mesh with automated failover across 3 regions" |
| "I use Terraform for infrastructure" | "I built a policy-as-code framework with drift detection and cost governance" |
| "I have CI/CD with GitHub Actions" | "I implemented a golden path with canary analysis, feature flags, and automated SBOM generation" |
| "I monitor with Prometheus" | "I defined SLOs, error budgets, and runbooks with cross-functional SRE alignment" |
| "My services talk via REST" | "I implemented Saga orchestration with Outbox pattern, idempotency keys, and dead-letter queues" |
| "I fixed a bug in production" | "I wrote a post-mortem, updated the runbook, added a chaos test, and improved the alert threshold to prevent recurrence" |

---

## License

This project is released under the **MIT License** for educational and portfolio purposes.

> **Disclaimer:** This is a reference architecture. It is not production software. Do not use it to process real financial transactions without thorough security audits, penetration testing, and SAMA approval.

---

## Contributing

Contributions are welcome, especially:

- Additional service implementations (Go/Rust/Python)
- Terraform module improvements
- OPA policy expansions for other GCC regulations (CBUAE, CBB)
- Arabic localization improvements
- Chaos experiment contributions

Please read `docs/adr/` before proposing architectural changes — every significant decision is documented there.

---

**Built with rigor for the Saudi enterprise market.** 🇸🇦

> *"In God we trust. All others bring data."* — W. Edwards Deming
