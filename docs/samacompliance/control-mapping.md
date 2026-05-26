# SAMA Cyber Security Framework Control Mapping

> This document maps the Saudi Central Bank (SAMA) Cyber Security Framework v1.0 controls to specific technical implementations in the Saudi Distributed Commerce & Settlement Fabric.

## Overview

| Domain | SAMA Controls | Implementation Status | Evidence Location |
|--------|--------------|----------------------|-------------------|
| Cyber Security Governance | 1.1.1 - 1.5.2 | Partial | `docs/adr/` |
| Cyber Security Risk Management | 2.1.1 - 2.4.2 | Implemented | `infrastructure/policies/opa/` |
| Cyber Security Operations | 3.1.1 - 3.6.2 | Implemented | Terraform, Kubernetes manifests |
| Cyber Security Architecture | 4.1.1 - 4.4.2 | Implemented | `docs/adr/`, Terraform modules |
| Cyber Security Monitoring | 5.1.1 - 5.3.2 | Implemented | Observability stack |

---

## Domain 3: Cyber Security Operations

### 3.1.1 - Asset Inventory
**Requirement**: Maintain an inventory of all assets and their classification.

**Implementation**:
- All AWS resources tagged with `ServiceName`, `Environment`, `DataClassification`, `Owner`
- Terraform enforces mandatory tags via OPA policy (`infrastructure/policies/sama_compliance.rego`)
- AWS Config rules auto-discover untagged resources

**Evidence**:
```bash
# Query all tagged resources
aws resourcegroupstaggingapi get-resources --tag-filters Key=DataClassification
```

### 3.1.2 - Data Classification
**Requirement**: Classify data based on sensitivity and apply appropriate controls.

**Implementation**:
| Classification | Color | Storage | Encryption | Access Control |
|----------------|-------|---------|------------|----------------|
| Public | Green | S3 standard | AES-256 | Public read |
| Internal | Yellow | EBS/S3 | AES-256 | IAM role-based |
| Confidential | Orange | RDS/S3 | KMS CMK | IRSA + Vault |
| Restricted | Red | RDS/S3 | KMS CMK + HSM | Vault dynamic creds |

**Evidence**: S3 bucket tags + Macie auto-discovery reports

### 3.2.1 - Network Security
**Requirement**: Implement network security controls to protect against unauthorized access.

**Implementation**:
- VPC with private subnets only (no public IPs on workloads)
- Security groups: zero `0.0.0.0/0` rules on sensitive ports
- VPC endpoints for all AWS services (no NAT Gateway for S3/ECR/CloudWatch)
- Istio AuthorizationPolicies enforce service-to-service ACLs

**Evidence**:
```hcl
# Terraform: VPC module
resource "aws_vpc_endpoint" "s3" { ... }
resource "aws_security_group_rule" "ingress" {
  cidr_blocks = [aws_vpc.main.cidr_block]  # Never 0.0.0.0/0
}
```

### 3.3.1 - Identity and Access Management
**Requirement**: Implement IAM controls with principle of least privilege.

**Implementation**:
- **No static AWS credentials**: IRSA (IAM Roles for Service Accounts) for all pods
- **No IAM users with console access**: SSO-only via AWS IAM Identity Center
- **Keycloak**: OAuth2/OIDC for application-level identity
- **Vault**: Dynamic database credentials with 1-hour TTL

**Evidence**: `infrastructure/terraform/modules/eks/` (IRSA roles), Vault audit logs

### 3.4.1 - Cryptography
**Requirement**: Implement encryption for data at rest and in transit.

**Implementation**:
| Layer | At Rest | In Transit |
|-------|---------|------------|
| EBS | KMS CMK | N/A |
| RDS Aurora | KMS CMK | TLS 1.3 |
| S3 | KMS CMK | TLS 1.3 |
| MSK | KMS CMK | TLS 1.3 + SASL/SCRAM |
| EKS Secrets | KMS CMK | N/A |
| Service Mesh | N/A | mTLS (Istio STRICT) |

**Evidence**: AWS Config rule `encrypted-volumes`, KMS CloudTrail logs

### 3.5.1 - Application Security
**Requirement**: Implement secure software development lifecycle (SSDLC).

**Implementation**:
- **SAST**: SonarQube in CI (0 critical issues gate)
- **DAST**: Trivy container scan + Snyk dependency check
- **SCA**: SBOM generation (CycloneDX) + Dependency-Track
- **Secrets scanning**: Trivy FS scan + GitHub secret scanning
- **Code signing**: Cosign image signatures

**Evidence**: `.github/workflows/golden-path.yml`

### 3.6.1 - Vulnerability Management
**Requirement**: Identify and remediate vulnerabilities in a timely manner.

**Implementation**:
- **Weekly patch cadence**: Automated PRs for base image updates
- **Critical CVE SLA**: 24 hours for Critical, 72 hours for High
- **Runtime detection**: Falco rules for container escape attempts

**Evidence**: Trivy scan reports in GitHub Security tab

---

## Domain 4: Cyber Security Architecture

### 4.1.1 - Defense in Depth
**Requirement**: Implement layered security controls.

**Implementation**:
```
Layer 1 (Edge): CloudFront + AWS WAF + Shield Advanced
Layer 2 (Network): VPC + private subnets + security groups + Istio mTLS
Layer 3 (Application): Keycloak RBAC + Istio AuthorizationPolicy
Layer 4 (Data): KMS encryption + Vault dynamic secrets + RDS encryption
Layer 5 (Monitoring): CloudTrail + Falco + Audit Service
```

### 4.2.1 - High Availability
**Requirement**: Ensure systems are resilient to failures.

**Implementation**:
- Multi-AZ EKS (3 AZs in me-south-1)
- Multi-AZ RDS Aurora with 1 reader
- Multi-AZ MSK with 3 brokers
- Istio circuit breakers + retries
- Automated chaos experiments (weekly)

**Evidence**: `scripts/chaos/weekly-experiments.yaml`

### 4.3.1 - Data Residency
**Requirement**: Ensure KSA customer data remains in KSA.

**Implementation**:
- All production workloads in `me-south-1` (Bahrain region serves KSA)
- S3 buckets with bucket policy denying cross-region replication
- RDS snapshots restricted to me-south-1
- Kafka topics: no MirrorMaker to non-KSA regions

**Evidence**: S3 bucket policies, RDS snapshot policies

---

## Domain 5: Cyber Security Monitoring

### 5.2.1 - Logging and Monitoring
**Requirement**: Collect and retain logs for audit and forensic purposes.

**Implementation**:
- **CloudTrail**: All AWS API calls retained for 7 years in Glacier
- **VPC Flow Logs**: Retained for 365 days in CloudWatch Logs
- **Application Logs**: Loki + S3 with structured JSON format
- **Audit Service**: Immutable append-only logs in Cassandra/S3 with Object Lock

**Evidence**: `infrastructure/terraform/modules/vpc/` (flow logs), Audit Service code

### 5.2.2 - Security Event Correlation
**Requirement**: Correlate security events across systems.

**Implementation**:
- Correlation IDs propagated across all Saga transactions
- OpenTelemetry traces link HTTP requests → Kafka messages → DB queries
- PagerDuty incident enrichment with trace IDs and log links

---

## Compliance Automation

### OPA Policy Enforcement
```bash
# Validate Terraform plan against SAMA policies
opa eval --data infrastructure/policies/sama_compliance.rego \
  --input tfplan.json "data.fabric.sama_compliance.violation"
```

### Monthly Compliance Scan
```yaml
# .github/workflows/sama-compliance.yml (recommended)
- name: Terraform Compliance Scan
  uses: bridgecrewio/checkov-action@master
  with:
    directory: infrastructure/terraform/
    framework: terraform
    output_format: sarif
```

---

## Audit Evidence Package

For SAMA audits, the following artifacts are produced quarterly:

1. **Terraform state exports** (infrastructure drift reports)
2. **Vulnerability scan reports** (Trivy + Snyk)
3. **Penetration test results** (external vendor)
4. **Chaos experiment results** (resilience validation)
5. **Incident post-mortems** (P1/P2 incidents)
6. **Access review logs** ( quarterly IAM/Keycloak access recertification)
