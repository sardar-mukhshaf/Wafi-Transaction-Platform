# ------------------------------------------------------------------------------
# OPA Policy: SAMA Cyber Security Framework v1.0 Compliance
# ------------------------------------------------------------------------------
# These policies enforce Saudi Central Bank (SAMA) requirements on AWS
# infrastructure provisioned via Terraform.
# ------------------------------------------------------------------------------

package fabric.sama_compliance

import future.keywords.if
import future.keywords.in

# Deny if S3 bucket is not encrypted
deny[msg] if {
    input.resource.aws_s3_bucket
    bucket := input.resource.aws_s3_bucket[_]
    not bucket.server_side_encryption_configuration
    msg := sprintf("S3 bucket '%s' must have server-side encryption enabled (SAMA 3.4.1)", [bucket.bucket])
}

# Deny if S3 bucket allows public access
deny[msg] if {
    input.resource.aws_s3_bucket_public_access_block
    block := input.resource.aws_s3_bucket_public_access_block[_]
    not block.block_public_acls
    msg := sprintf("S3 bucket '%s' must block public ACLs (SAMA 3.3.1)", [block.bucket])
}

deny[msg] if {
    input.resource.aws_s3_bucket_public_access_block
    block := input.resource.aws_s3_bucket_public_access_block[_]
    not block.block_public_policy
    msg := sprintf("S3 bucket '%s' must block public policies (SAMA 3.3.1)", [block.bucket])
}

# Deny if RDS instance is not encrypted
deny[msg] if {
    input.resource.aws_rds_cluster
    cluster := input.resource.aws_rds_cluster[_]
    not cluster.storage_encrypted
    msg := sprintf("RDS cluster '%s' must have storage encryption enabled (SAMA 3.4.1)", [cluster.cluster_identifier])
}

# Deny if security group allows 0.0.0.0/0 on sensitive ports
deny[msg] if {
    input.resource.aws_security_group_rule
    rule := input.resource.aws_security_group_rule[_]
    rule.cidr_blocks[_] == "0.0.0.0/0"
    sensitive_ports := [22, 3306, 5432, 6379, 9200, 27017]
    rule.from_port in sensitive_ports
    msg := sprintf("Security group rule allows 0.0.0.0/0 on port %d (SAMA 3.2.1)", [rule.from_port])
}

# Deny if resource is missing required tags
deny[msg] if {
    input.resource.aws_instance
    instance := input.resource.aws_instance[_]
    not instance.tags["Environment"]
    msg := sprintf("EC2 instance '%s' must have Environment tag (SAMA governance)", [instance.instance_type])
}

deny[msg] if {
    input.resource.aws_instance
    instance := input.resource.aws_instance[_]
    not instance.tags["DataClassification"]
    msg := sprintf("EC2 instance '%s' must have DataClassification tag (SAMA 3.1.2)", [instance.instance_type])
}

# Deny if KMS key rotation is not enabled
deny[msg] if {
    input.resource.aws_kms_key
    key := input.resource.aws_kms_key[_]
    not key.enable_key_rotation
    msg := "KMS keys must have automatic key rotation enabled (SAMA 3.4.1)"
}

# Deny if VPC Flow Logs are not enabled
deny[msg] if {
    input.resource.aws_vpc
    vpc := input.resource.aws_vpc[_]
    not input.resource.aws_flow_log
    msg := sprintf("VPC '%s' must have VPC Flow Logs enabled (SAMA 5.2.1)", [vpc.cidr_block])
}

# Deny if EKS cluster public endpoint is enabled without CIDR restrictions
deny[msg] if {
    input.resource.aws_eks_cluster
    cluster := input.resource.aws_eks_cluster[_]
    cluster.vpc_config[0].endpoint_public_access
    count(cluster.vpc_config[0].public_access_cidrs) == 0
    msg := sprintf("EKS cluster '%s' public endpoint must have CIDR restrictions (SAMA 3.2.1)", [cluster.name])
}

# Deny if IAM user has console access (force IAM roles only)
deny[msg] if {
    input.resource.aws_iam_user_login_profile
    msg := "IAM console access via login profiles is not allowed. Use IAM roles with SSO (SAMA 3.3.1)"
}

# Allow if no deny rules match
allow := true if {
    count(deny) == 0
}

# Main evaluation result
violation := {
    "allow": allow,
    "deny": deny,
}
