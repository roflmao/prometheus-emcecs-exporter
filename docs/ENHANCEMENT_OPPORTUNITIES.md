# Prometheus ECS Exporter Enhancement Opportunities

**Document Version**: 1.0
**Date**: 2024-11-19
**Target Platform**: Dell ObjectScale 4.1 / ECS 3.9 API

---

## Executive Summary

ObjectScale 4.1 (ECS 3.9 API) introduces new capabilities that could significantly enhance the Prometheus exporter's monitoring coverage. This document outlines three major enhancement areas with concrete metric examples and implementation guidance.

**Enhancement Areas**:
1. **Billing & Cost Monitoring** - Track storage costs and usage patterns
2. **Enhanced System Metrics** - Deeper operational visibility
3. **Multi-Tenant IAM Monitoring** - Identity and access tracking

---

## 1. Billing & Cost Monitoring

### Overview

The new Billing Service API provides detailed cost and usage information at namespace and bucket levels, enabling FinOps capabilities for ObjectScale deployments.

### API Endpoints Available

- `POST /object/billing/buckets/{namespace}/info` - Get bucket billing details
- `POST /object/billing/buckets/{namespace}/sample` - Get billing samples over time intervals
- `POST /object/billing/namespace/info` - Get namespace-level billing
- `POST /object/billing/namespace/sample` - Get namespace billing samples over intervals

### Concrete Metrics Examples

#### Example 1: Bucket Storage Cost Tracking
```prometheus
# METRIC: emcecs_bucket_storage_bytes
# TYPE: gauge
# HELP: Total storage used by bucket in bytes
emcecs_bucket_storage_bytes{namespace="tenant1",bucket="logs",storage_class="standard"} 52428800000

# METRIC: emcecs_bucket_storage_cost_usd
# TYPE: gauge
# HELP: Estimated monthly storage cost in USD for bucket
emcecs_bucket_storage_cost_usd{namespace="tenant1",bucket="logs",storage_class="standard"} 1.20

# Use Case: Track which buckets are consuming the most storage budget
# Query: topk(10, emcecs_bucket_storage_cost_usd)
```

**Implementation Details**:
- Call billing API every 5-15 minutes (configurable)
- Cache results to reduce API calls
- Parse billing response for storage size and cost calculations
- Support multiple cost models (per GB, per object, etc.)

#### Example 2: Namespace-Level Usage Tracking
```prometheus
# METRIC: emcecs_namespace_total_objects
# TYPE: gauge
# HELP: Total number of objects in namespace
emcecs_namespace_total_objects{namespace="tenant1",vdc="vdc1"} 1523487

# METRIC: emcecs_namespace_total_bytes
# TYPE: gauge
# HELP: Total bytes stored in namespace
emcecs_namespace_total_bytes{namespace="tenant1",vdc="vdc1"} 524288000000

# METRIC: emcecs_namespace_ingress_bytes_total
# TYPE: counter
# HELP: Total bytes ingressed to namespace
emcecs_namespace_ingress_bytes_total{namespace="tenant1",vdc="vdc1"} 1073741824000

# Use Case: Multi-tenant chargeback and quota monitoring
# Query: sum(rate(emcecs_namespace_ingress_bytes_total[5m])) by (namespace)
```

**Implementation Details**:
- Poll namespace billing endpoint per configured interval
- Track ingress/egress for bandwidth billing
- Calculate growth rates for capacity planning
- Support namespace-level quotas and alerts

#### Example 3: Time-Series Cost Analysis
```prometheus
# METRIC: emcecs_bucket_cost_sample_usd
# TYPE: gauge
# HELP: Bucket cost sample at specific timestamp for trend analysis
emcecs_bucket_cost_sample_usd{namespace="tenant1",bucket="logs",timestamp="1700000000"} 1.15
emcecs_bucket_cost_sample_usd{namespace="tenant1",bucket="logs",timestamp="1700086400"} 1.18
emcecs_bucket_cost_sample_usd{namespace="tenant1",bucket="logs",timestamp="1700172800"} 1.20

# METRIC: emcecs_bucket_cost_trend_percent
# TYPE: gauge
# HELP: Percentage cost change over last 24 hours
emcecs_bucket_cost_trend_percent{namespace="tenant1",bucket="logs"} 4.35

# Use Case: Detect unexpected cost increases
# Query: emcecs_bucket_cost_trend_percent > 10
```

**Implementation Details**:
- Use billing sample API with time ranges
- Store historical data points for trend calculation
- Calculate cost deltas and growth percentages
- Generate alerts on cost anomalies

#### Example 4: Operations-Based Billing
```prometheus
# METRIC: emcecs_bucket_api_operations_total
# TYPE: counter
# HELP: Total API operations performed on bucket
emcecs_bucket_api_operations_total{namespace="tenant1",bucket="logs",operation="GET"} 15234879
emcecs_bucket_api_operations_total{namespace="tenant1",bucket="logs",operation="PUT"} 523487
emcecs_bucket_api_operations_total{namespace="tenant1",bucket="logs",operation="DELETE"} 12345

# METRIC: emcecs_bucket_operation_cost_usd
# TYPE: gauge
# HELP: Cost of API operations for bucket
emcecs_bucket_operation_cost_usd{namespace="tenant1",bucket="logs",operation="GET"} 0.076
emcecs_bucket_operation_cost_usd{namespace="tenant1",bucket="logs",operation="PUT"} 0.262

# Use Case: Track API operation costs for S3-compatible pricing
# Query: sum(rate(emcecs_bucket_api_operations_total[1h])) by (operation, namespace)
```

**Implementation Details**:
- Parse operation counts from billing data
- Apply cost-per-operation pricing model
- Track read vs write operation ratios
- Support tiered pricing based on volume

#### Example 5: Data Transfer Cost Tracking
```prometheus
# METRIC: emcecs_namespace_egress_bytes_total
# TYPE: counter
# HELP: Total bytes egressed from namespace
emcecs_namespace_egress_bytes_total{namespace="tenant1",destination="internet"} 10737418240
emcecs_namespace_egress_bytes_total{namespace="tenant1",destination="internal"} 53687091200

# METRIC: emcecs_namespace_bandwidth_cost_usd
# TYPE: gauge
# HELP: Data transfer cost for namespace
emcecs_namespace_bandwidth_cost_usd{namespace="tenant1",direction="egress",destination="internet"} 0.85
emcecs_namespace_bandwidth_cost_usd{namespace="tenant1",direction="ingress"} 0.00

# Use Case: Monitor data transfer costs for cloud parity pricing
# Query: sum(emcecs_namespace_bandwidth_cost_usd) by (namespace, destination)
```

**Implementation Details**:
- Track ingress/egress separately
- Differentiate internal vs internet traffic
- Apply tiered bandwidth pricing
- Calculate monthly bandwidth projections

---

## 2. Enhanced System Metrics

### Overview

ObjectScale 4.1 exposes additional operational metrics through enhanced monitoring APIs, providing deeper visibility into system health and performance.

### API Endpoints Available

- Enhanced `/dashboard/zones/localzone` - Additional system metrics
- Improved node statistics via port 9101
- New monitoring events API
- Enhanced capacity reporting

### Concrete Metrics Examples

#### Example 1: Advanced Capacity Metrics
```prometheus
# METRIC: emcecs_storage_efficiency_ratio
# TYPE: gauge
# HELP: Storage efficiency ratio (logical/physical)
emcecs_storage_efficiency_ratio{vdc="vdc1",storage_pool="sp1"} 2.34

# METRIC: emcecs_storage_compression_ratio
# TYPE: gauge
# HELP: Data compression ratio achieved
emcecs_storage_compression_ratio{vdc="vdc1",storage_pool="sp1"} 1.87

# METRIC: emcecs_storage_deduplication_ratio
# TYPE: gauge
# HELP: Data deduplication ratio achieved
emcecs_storage_deduplication_ratio{vdc="vdc1",storage_pool="sp1"} 1.25

# METRIC: emcecs_storage_effective_capacity_bytes
# TYPE: gauge
# HELP: Effective capacity after compression and deduplication
emcecs_storage_effective_capacity_bytes{vdc="vdc1",storage_pool="sp1"} 5497558138880

# Use Case: Track storage optimization effectiveness
# Query: avg(emcecs_storage_efficiency_ratio) by (storage_pool)
```

**Implementation Details**:
- Parse enhanced capacity API responses
- Calculate efficiency ratios from raw/cooked data
- Track compression/dedup effectiveness
- Monitor optimization trends over time

#### Example 2: Object Lock & Compliance Metrics
```prometheus
# METRIC: emcecs_bucket_locked_objects_total
# TYPE: gauge
# HELP: Total number of locked objects in bucket
emcecs_bucket_locked_objects_total{namespace="compliance",bucket="archives",lock_type="retention"} 15234
emcecs_bucket_locked_objects_total{namespace="compliance",bucket="archives",lock_type="legal_hold"} 487

# METRIC: emcecs_bucket_lock_remaining_days
# TYPE: gauge
# HELP: Minimum days remaining until oldest lock expires
emcecs_bucket_lock_remaining_days{namespace="compliance",bucket="archives",lock_type="retention"} 1825

# METRIC: emcecs_bucket_compliance_violations_total
# TYPE: counter
# HELP: Total compliance violations detected
emcecs_bucket_compliance_violations_total{namespace="compliance",bucket="archives",violation_type="early_deletion_attempt"} 3

# Use Case: Monitor compliance and retention policies
# Query: emcecs_bucket_lock_remaining_days < 30
```

**Implementation Details**:
- Query bucket lock configuration APIs
- Track retention periods and legal holds
- Monitor compliance events
- Alert on policy violations

#### Example 3: Replication Performance Metrics
```prometheus
# METRIC: emcecs_replication_lag_seconds
# TYPE: gauge
# HELP: Replication lag in seconds between sites
emcecs_replication_lag_seconds{source_vdc="vdc1",target_vdc="vdc2",rg="rg1"} 12.5

# METRIC: emcecs_replication_throughput_bytes_per_sec
# TYPE: gauge
# HELP: Current replication throughput in bytes per second
emcecs_replication_throughput_bytes_per_sec{source_vdc="vdc1",target_vdc="vdc2",rg="rg1"} 104857600

# METRIC: emcecs_replication_queue_depth
# TYPE: gauge
# HELP: Number of objects waiting to be replicated
emcecs_replication_queue_depth{source_vdc="vdc1",target_vdc="vdc2",rg="rg1"} 523

# METRIC: emcecs_replication_failed_objects_total
# TYPE: counter
# HELP: Total failed replication attempts
emcecs_replication_failed_objects_total{source_vdc="vdc1",target_vdc="vdc2",rg="rg1"} 7

# Use Case: Monitor geo-replication health and performance
# Query: emcecs_replication_lag_seconds > 300
```

**Implementation Details**:
- Parse enhanced replication group data
- Calculate lag from timestamps
- Monitor queue depths and backlogs
- Track replication failures and retries

#### Example 4: Erasure Coding & Data Protection
```prometheus
# METRIC: emcecs_erasure_coding_scheme
# TYPE: gauge
# HELP: Erasure coding scheme (k+m) encoded as k*100+m
emcecs_erasure_coding_scheme{storage_pool="sp1"} 1202

# METRIC: emcecs_data_protection_overhead_ratio
# TYPE: gauge
# HELP: Storage overhead ratio for data protection
emcecs_data_protection_overhead_ratio{storage_pool="sp1"} 1.166

# METRIC: emcecs_chunk_rebuild_operations_total
# TYPE: counter
# HELP: Total chunk rebuild operations performed
emcecs_chunk_rebuild_operations_total{storage_pool="sp1",node="node1"} 234

# METRIC: emcecs_chunk_rebuild_duration_seconds
# TYPE: histogram
# HELP: Duration of chunk rebuild operations
emcecs_chunk_rebuild_duration_seconds_bucket{storage_pool="sp1",le="10"} 150
emcecs_chunk_rebuild_duration_seconds_bucket{storage_pool="sp1",le="30"} 220
emcecs_chunk_rebuild_duration_seconds_bucket{storage_pool="sp1",le="+Inf"} 234

# Use Case: Monitor data protection and recovery performance
# Query: rate(emcecs_chunk_rebuild_operations_total[5m])
```

**Implementation Details**:
- Parse storage pool configurations
- Track EC scheme details (k+m values)
- Monitor rebuild operations and performance
- Calculate protection overhead

#### Example 5: S3 Multipart Upload Tracking
```prometheus
# METRIC: emcecs_s3_multipart_uploads_active
# TYPE: gauge
# HELP: Number of active multipart uploads
emcecs_s3_multipart_uploads_active{namespace="uploads",bucket="large-files"} 47

# METRIC: emcecs_s3_multipart_uploads_abandoned_total
# TYPE: counter
# HELP: Total abandoned multipart uploads
emcecs_s3_multipart_uploads_abandoned_total{namespace="uploads",bucket="large-files"} 12

# METRIC: emcecs_s3_multipart_incomplete_bytes
# TYPE: gauge
# HELP: Storage consumed by incomplete multipart uploads
emcecs_s3_multipart_incomplete_bytes{namespace="uploads",bucket="large-files"} 5368709120

# METRIC: emcecs_s3_multipart_upload_duration_seconds
# TYPE: histogram
# HELP: Duration of completed multipart uploads
emcecs_s3_multipart_upload_duration_seconds_bucket{namespace="uploads",bucket="large-files",le="60"} 25
emcecs_s3_multipart_upload_duration_seconds_bucket{namespace="uploads",bucket="large-files",le="300"} 40
emcecs_s3_multipart_upload_duration_seconds_bucket{namespace="uploads",bucket="large-files",le="+Inf"} 47

# Use Case: Track large file uploads and identify incomplete transfers
# Query: emcecs_s3_multipart_incomplete_bytes > 10737418240
```

**Implementation Details**:
- Query S3 multipart upload status APIs
- Track active and incomplete uploads
- Monitor abandoned uploads for cleanup
- Measure upload performance

---

## 3. Multi-Tenant IAM Monitoring

### Overview

ObjectScale 4.1's enhanced IAM integration provides detailed identity and access management metrics, enabling security monitoring and compliance tracking across multi-tenant deployments.

### API Endpoints Available

- IAM Service API - User and role management
- STS Service API - Temporary credentials
- Authentication Provider API - Auth backend status
- User Management API - Enhanced user tracking

### Concrete Metrics Examples

#### Example 1: IAM User Activity Tracking
```prometheus
# METRIC: emcecs_iam_users_total
# TYPE: gauge
# HELP: Total IAM users per namespace
emcecs_iam_users_total{namespace="tenant1",user_type="human"} 47
emcecs_iam_users_total{namespace="tenant1",user_type="service_account"} 23

# METRIC: emcecs_iam_users_active_total
# TYPE: gauge
# HELP: IAM users with activity in last 90 days
emcecs_iam_users_active_total{namespace="tenant1",user_type="human"} 38
emcecs_iam_users_active_total{namespace="tenant1",user_type="service_account"} 23

# METRIC: emcecs_iam_users_inactive_days
# TYPE: gauge
# HELP: Days since user last activity
emcecs_iam_users_inactive_days{namespace="tenant1",user="john.doe",user_type="human"} 127

# METRIC: emcecs_iam_credential_age_days
# TYPE: gauge
# HELP: Age of IAM credentials in days
emcecs_iam_credential_age_days{namespace="tenant1",user="service-app1",credential_type="access_key"} 243

# Use Case: Identify dormant accounts and aged credentials
# Query: emcecs_iam_users_inactive_days > 90
```

**Implementation Details**:
- Query IAM user list and metadata
- Track last login/activity timestamps
- Calculate credential age
- Monitor credential rotation compliance

#### Example 2: Authentication Provider Health
```prometheus
# METRIC: emcecs_auth_provider_available
# TYPE: gauge
# HELP: Authentication provider availability (1=up, 0=down)
emcecs_auth_provider_available{provider="ldap-primary",namespace="tenant1"} 1
emcecs_auth_provider_available{provider="ldap-secondary",namespace="tenant1"} 1
emcecs_auth_provider_available{provider="saml-okta",namespace="tenant2"} 1

# METRIC: emcecs_auth_provider_response_time_seconds
# TYPE: gauge
# HELP: Authentication provider response time
emcecs_auth_provider_response_time_seconds{provider="ldap-primary",namespace="tenant1"} 0.045

# METRIC: emcecs_auth_provider_failures_total
# TYPE: counter
# HELP: Total authentication failures by provider
emcecs_auth_provider_failures_total{provider="ldap-primary",namespace="tenant1",reason="timeout"} 3
emcecs_auth_provider_failures_total{provider="ldap-primary",namespace="tenant1",reason="invalid_credentials"} 127

# METRIC: emcecs_auth_provider_users_count
# TYPE: gauge
# HELP: Number of users from this provider
emcecs_auth_provider_users_count{provider="ldap-primary",namespace="tenant1"} 234

# Use Case: Monitor auth provider health and failover scenarios
# Query: emcecs_auth_provider_available == 0
```

**Implementation Details**:
- Poll authentication provider status API
- Test provider connectivity and response time
- Track authentication success/failure rates
- Monitor user distribution across providers

#### Example 3: STS Token Usage Metrics
```prometheus
# METRIC: emcecs_sts_tokens_active
# TYPE: gauge
# HELP: Number of active STS tokens
emcecs_sts_tokens_active{namespace="tenant1",session_type="assume_role"} 142
emcecs_sts_tokens_active{namespace="tenant1",session_type="federated"} 23

# METRIC: emcecs_sts_tokens_issued_total
# TYPE: counter
# HELP: Total STS tokens issued
emcecs_sts_tokens_issued_total{namespace="tenant1",session_type="assume_role"} 15234
emcecs_sts_tokens_issued_total{namespace="tenant1",session_type="federated"} 1523

# METRIC: emcecs_sts_token_remaining_seconds
# TYPE: gauge
# HELP: Seconds remaining for shortest-lived active token
emcecs_sts_token_remaining_seconds{namespace="tenant1",session_type="assume_role"} 1800

# METRIC: emcecs_sts_token_duration_seconds
# TYPE: histogram
# HELP: Duration of issued STS tokens
emcecs_sts_token_duration_seconds_bucket{namespace="tenant1",le="3600"} 8500
emcecs_sts_token_duration_seconds_bucket{namespace="tenant1",le="7200"} 14200
emcecs_sts_token_duration_seconds_bucket{namespace="tenant1",le="+Inf"} 15234

# Use Case: Monitor temporary credential usage and security policies
# Query: rate(emcecs_sts_tokens_issued_total[5m]) by (namespace)
```

**Implementation Details**:
- Query STS service for active sessions
- Track token issuance rates
- Monitor token lifetimes and expiration
- Alert on unusual token usage patterns

#### Example 4: IAM Role & Policy Compliance
```prometheus
# METRIC: emcecs_iam_roles_total
# TYPE: gauge
# HELP: Total IAM roles defined
emcecs_iam_roles_total{namespace="tenant1"} 23

# METRIC: emcecs_iam_policies_total
# TYPE: gauge
# HELP: Total IAM policies defined
emcecs_iam_policies_total{namespace="tenant1",policy_type="managed"} 45
emcecs_iam_policies_total{namespace="tenant1",policy_type="inline"} 78

# METRIC: emcecs_iam_policy_attachments_total
# TYPE: gauge
# HELP: Total policy attachments to users/roles
emcecs_iam_policy_attachments_total{namespace="tenant1",attached_to="user"} 89
emcecs_iam_policy_attachments_total{namespace="tenant1",attached_to="role"} 67

# METRIC: emcecs_iam_overprivileged_entities
# TYPE: gauge
# HELP: Number of entities with admin-level permissions
emcecs_iam_overprivileged_entities{namespace="tenant1",entity_type="user"} 5
emcecs_iam_overprivileged_entities{namespace="tenant1",entity_type="role"} 2

# Use Case: Monitor IAM configuration and security posture
# Query: emcecs_iam_overprivileged_entities > 10
```

**Implementation Details**:
- Enumerate IAM roles and policies
- Parse policy documents for permission analysis
- Identify overly permissive policies
- Track policy attachment patterns

#### Example 5: Multi-Tenant Access Metrics
```prometheus
# METRIC: emcecs_namespace_access_requests_total
# TYPE: counter
# HELP: Total access requests per namespace
emcecs_namespace_access_requests_total{namespace="tenant1",result="success"} 1523487
emcecs_namespace_access_requests_total{namespace="tenant1",result="denied"} 1234

# METRIC: emcecs_namespace_cross_tenant_requests_total
# TYPE: counter
# HELP: Cross-tenant access attempts
emcecs_namespace_cross_tenant_requests_total{source_namespace="tenant1",target_namespace="tenant2",result="denied"} 5

# METRIC: emcecs_namespace_access_denied_by_reason
# TYPE: counter
# HELP: Access denials by reason
emcecs_namespace_access_denied_by_reason{namespace="tenant1",reason="insufficient_permissions"} 892
emcecs_namespace_access_denied_by_reason{namespace="tenant1",reason="expired_token"} 234
emcecs_namespace_access_denied_by_reason{namespace="tenant1",reason="invalid_signature"} 108

# METRIC: emcecs_namespace_privileged_operations_total
# TYPE: counter
# HELP: Privileged operations performed
emcecs_namespace_privileged_operations_total{namespace="tenant1",operation="DeleteBucket"} 12
emcecs_namespace_privileged_operations_total{namespace="tenant1",operation="PutBucketPolicy"} 34

# Use Case: Security monitoring and access pattern analysis
# Query: rate(emcecs_namespace_cross_tenant_requests_total[5m]) > 0
```

**Implementation Details**:
- Parse access logs and audit trails
- Track authorization decisions
- Monitor cross-tenant access patterns
- Alert on suspicious access patterns

---

## Implementation Roadmap

### Phase 1: Billing & Cost Monitoring (Estimated Effort: 2-3 weeks)

**Priority**: High - Enables FinOps capabilities

**Tasks**:
1. Implement billing API client
   - Add billing service endpoints to ecsclient.go
   - Create data structures for billing responses
   - Implement caching layer for billing data

2. Create billing collector
   - Namespace-level billing metrics
   - Bucket-level billing metrics
   - Time-series sampling support

3. Testing & validation
   - Unit tests for billing calculations
   - Integration tests with test namespace
   - Performance testing (API call frequency)

**Deliverables**:
- 15-20 new billing-related metrics
- Configuration options for billing polling intervals
- Documentation for cost monitoring use cases

### Phase 2: Enhanced System Metrics (Estimated Effort: 3-4 weeks)

**Priority**: Medium - Improves operational visibility

**Tasks**:
1. Extend capacity metrics
   - Parse enhanced capacity API responses
   - Add efficiency/compression/dedup metrics
   - Implement effective capacity calculations

2. Add compliance metrics
   - Object lock status tracking
   - Retention period monitoring
   - Legal hold tracking

3. Enhance replication metrics
   - Replication lag calculations
   - Queue depth monitoring
   - Failure tracking and alerting

4. Add data protection metrics
   - Erasure coding scheme detection
   - Rebuild operation tracking
   - Protection overhead calculations

5. Implement S3 multipart tracking
   - Active upload monitoring
   - Incomplete upload detection
   - Performance histogram collection

**Deliverables**:
- 25-30 new operational metrics
- Enhanced dashboard examples for Grafana
- Alerting rule examples

### Phase 3: Multi-Tenant IAM Monitoring (Estimated Effort: 2-3 weeks)

**Priority**: Medium - Enables security monitoring

**Tasks**:
1. Implement IAM API client
   - Add IAM service endpoints
   - Add STS service endpoints
   - Create auth provider monitoring

2. Create IAM collector
   - User activity tracking
   - Credential age monitoring
   - Auth provider health checks

3. Add security metrics
   - Token usage tracking
   - Access pattern monitoring
   - Policy compliance checks

**Deliverables**:
- 20-25 new IAM/security metrics
- Security dashboard examples
- Compliance monitoring guidelines

### Phase 4: Testing & Documentation (Estimated Effort: 1-2 weeks)

**Priority**: High - Ensures quality and adoption

**Tasks**:
1. Comprehensive testing
   - Unit test coverage >80%
   - Integration tests with ObjectScale 4.1
   - Performance benchmarking

2. Documentation
   - Update README with new features
   - Create metric reference guide
   - Write Grafana dashboard examples
   - Document configuration options

3. Example dashboards
   - Cost monitoring dashboard
   - Security monitoring dashboard
   - Enhanced operations dashboard

**Deliverables**:
- Complete test suite
- User documentation
- 3-5 example Grafana dashboards

---

## Technical Considerations

### API Rate Limiting

**Concern**: Increased API calls may impact ObjectScale performance

**Mitigation**:
- Implement configurable polling intervals (default: 5 minutes for billing)
- Use caching layers for frequently accessed data
- Batch API requests where possible
- Add circuit breaker pattern for API failures
- Monitor exporter's own resource usage

### Backward Compatibility

**Concern**: New features should not break existing deployments

**Strategy**:
- Make all new collectors opt-in via configuration
- Detect API version and enable features dynamically
- Gracefully handle missing APIs (ECS 3.6 vs 4.1)
- Maintain existing metric names and labels
- Version configuration file format

### Performance Impact

**Metrics Collection**:
- Current exporter: ~50 metrics per scrape
- With enhancements: ~150-200 metrics per scrape
- Estimated scrape duration: 5-10 seconds (from 2-3 seconds)

**Memory Usage**:
- Current: ~50MB RSS
- Estimated with enhancements: ~100-150MB RSS
- Mitigation: Implement metric cardinality limits

**ObjectScale API Load**:
- Additional API calls per scrape: 10-15
- Estimated API response time: 100-500ms each
- Total additional load: 1-7 seconds per scrape interval

### Configuration Design

```yaml
# Example configuration for new features
billing:
  enabled: true
  poll_interval: 300s  # 5 minutes
  namespaces:
    - tenant1
    - tenant2
  include_samples: true
  sample_interval: 3600s  # 1 hour

enhanced_metrics:
  enabled: true
  capacity_details: true
  replication_lag: true
  erasure_coding: true
  multipart_uploads: true

iam_monitoring:
  enabled: false  # Opt-in for security reasons
  poll_interval: 600s  # 10 minutes
  track_inactive_users: true
  inactive_threshold_days: 90
  monitor_sts_tokens: true
  auth_providers:
    - ldap-primary
    - saml-okta
```

---

## Benefits & ROI

### Cost Optimization Benefits

**Visibility**:
- Track storage costs per tenant/bucket
- Identify cost optimization opportunities
- Enable accurate chargeback/showback

**Estimated Savings**:
- 10-20% reduction in storage costs through visibility
- Better capacity planning reducing over-provisioning
- Automated alerting on cost anomalies

### Operational Benefits

**Improved Monitoring**:
- Deeper system health visibility
- Proactive issue detection
- Better capacity planning

**Reduced MTTR**:
- Faster troubleshooting with detailed metrics
- Better root cause analysis
- Estimated 30-40% reduction in incident resolution time

### Security Benefits

**Compliance**:
- Audit trail for IAM operations
- Credential rotation monitoring
- Access pattern anomaly detection

**Risk Reduction**:
- Early detection of security issues
- Dormant account identification
- Overprivileged entity detection

---

## Example Grafana Dashboards

### Dashboard 1: Cost Monitoring

**Panels**:
1. Total Monthly Cost by Namespace (gauge)
2. Top 10 Expensive Buckets (table)
3. Storage Cost Trend (time series)
4. Cost by Storage Class (pie chart)
5. Bandwidth Cost Breakdown (bar chart)

**Alerts**:
- Cost increase >20% day-over-day
- Monthly budget threshold exceeded
- Unexpected billing spikes

### Dashboard 2: Multi-Tenant Overview

**Panels**:
1. Storage by Namespace (stacked area)
2. Active Users per Tenant (bar chart)
3. API Operations by Namespace (time series)
4. Cross-Tenant Access Attempts (stat)
5. IAM Health by Tenant (status grid)

**Alerts**:
- Namespace approaching quota
- Unusual cross-tenant access
- Auth provider failures

### Dashboard 3: Security Monitoring

**Panels**:
1. Active IAM Users (stat)
2. Inactive Accounts >90 Days (table)
3. Aged Credentials (gauge)
4. Authentication Failures by Provider (time series)
5. Privileged Operations Log (table)

**Alerts**:
- New overprivileged user detected
- Auth failure rate >5%
- Suspicious access patterns

---

## Conclusion

ObjectScale 4.1's new APIs provide significant opportunities to enhance the Prometheus exporter with:

- **60-70 new metrics** across billing, operations, and IAM
- **3-5x increase** in monitoring coverage
- **FinOps capabilities** for cost optimization
- **Enhanced security** monitoring for compliance

The proposed implementation is backward-compatible, opt-in, and provides clear ROI through cost savings, operational improvements, and risk reduction.

**Recommended Next Steps**:
1. Validate API access with ObjectScale 4.1 test instance
2. Prioritize Phase 1 (Billing) for immediate FinOps value
3. Create POC with 5-10 key metrics from each category
4. Gather community feedback on metric names and labels
5. Develop comprehensive implementation plan

---

## References

- ObjectScale 4.1 REST API Documentation: `/ObjectScale_4.1_REST_API/`
- Current Exporter Code: `/pkg/ecsclient/ecsclient.go`
- Prometheus Best Practices: https://prometheus.io/docs/practices/naming/
- Grafana Dashboard Examples: https://grafana.com/grafana/dashboards/

---

**Document Status**: Draft for Review
**Feedback**: Please submit issues or PRs to the prometheus-emcecs-exporter repository
