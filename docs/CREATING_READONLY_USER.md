# Creating a Read-Only Monitoring User for ObjectScale/ECS

**Document Version**: 1.0
**Date**: 2024-11-19
**Applies To**: Dell ObjectScale 4.1 / ECS 3.x

---

## Overview

The Prometheus ECS exporter requires authenticated access to the ObjectScale/ECS Management API. For security best practices, you should create a dedicated **read-only user** specifically for monitoring purposes.

---

## Method 1: Using the Web UI (Recommended)

This is the easiest and most straightforward method.

### Steps:

1. **Access the ECS/ObjectScale Management UI**
   ```
   https://<your-objectscale-host>:443
   ```

2. **Login as Administrator**
   - Use your admin credentials

3. **Navigate to User Management**
   - Go to: **Manage** → **Users** → **Management Users**
   - Or: **Settings** → **Users** → **Local Users**

4. **Create New Management User**
   - Click **"Create Management User"** or **"Add User"**

5. **Configure User Settings**:
   ```
   Username: prometheus-monitor
   Password: <generate strong password>
   Description: Read-only user for Prometheus monitoring
   ```

6. **Assign Role**:
   - Select role: **"Monitor"** or **"System Monitor"**
   - **DO NOT** select "System Administrator" or "System Auditor"

7. **Save and Test**
   - Click **"Save"**
   - Test authentication using curl (see Testing section below)

### Available Roles:

| Role | Permissions | Suitable for Monitoring? |
|------|-------------|--------------------------|
| **System Monitor** | Read-only access to all monitoring APIs | ✅ **YES - Recommended** |
| System Administrator | Full read/write access | ❌ NO - Too privileged |
| System Auditor | Read-only + audit logs | ⚠️ OK but overprivileged |
| Namespace Administrator | Namespace-specific admin | ❌ NO - Cannot access cluster metrics |

---

## Method 2: Using the REST API

If you prefer automation or don't have UI access, use the Management API.

### Prerequisites:
- Admin credentials
- Access to ObjectScale Management API (port 4443)

### Create User via API:

```bash
# 1. Get auth token
TOKEN=$(curl -k -u "admin:adminpassword" \
  -X GET "https://<objectscale-host>:4443/login" \
  | grep -oP 'X-SDS-AUTH-TOKEN: \K[^"]+')

# 2. Create management user
curl -k -X POST \
  -H "X-SDS-AUTH-TOKEN: ${TOKEN}" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  "https://<objectscale-host>:4443/vdc/users/mgmt" \
  -d '{
    "user": "prometheus-monitor",
    "password": "YourStrongPassword123!",
    "isSystemAdmin": false,
    "isSystemMonitor": true,
    "isSecurityAdmin": false
  }'
```

### API Endpoint Details:

**Endpoint**: `POST /vdc/users/mgmt`

**Request Body**:
```json
{
  "user": "prometheus-monitor",
  "password": "your-secure-password",
  "isSystemAdmin": false,
  "isSystemMonitor": true,
  "isSecurityAdmin": false
}
```

**Response**:
```json
{
  "user": "prometheus-monitor",
  "created": "2024-11-19T10:00:00Z",
  "id": "urn:storageos:VirtualDataCenterData:...",
  "isSystemMonitor": true
}
```

---

## Method 3: Using ECS CLI (if available)

Some ECS deployments have CLI tools:

```bash
# SSH to ECS node (if you have access)
ssh admin@<ecs-node>

# Create monitoring user
viprfs -c /opt/storageos/conf/coordinatorsvc-conf.xml create-user \
  --user prometheus-monitor \
  --password YourStrongPassword123! \
  --role SYSTEM_MONITOR
```

---

## Testing the Monitoring User

### Test 1: Basic Authentication
```bash
# Test login and get token
curl -k -u "prometheus-monitor:YourPassword" \
  -X GET "https://<objectscale-host>:4443/login"

# Expected output: X-SDS-AUTH-TOKEN header with token value
```

### Test 2: Access Dashboard API (Read)
```bash
# Get token first
TOKEN=$(curl -k -u "prometheus-monitor:YourPassword" \
  -X GET "https://<objectscale-host>:4443/login" \
  | grep -oP 'X-SDS-AUTH-TOKEN: \K[^"]+')

# Test dashboard access
curl -k -H "X-SDS-AUTH-TOKEN: ${TOKEN}" \
  -H "Accept: application/json" \
  "https://<objectscale-host>:4443/dashboard/zones/localzone" \
  | jq '.'

# Expected: JSON with cluster state information
```

### Test 3: Verify Read-Only (Should Fail)
```bash
# Try to create a bucket (should fail)
curl -k -H "X-SDS-AUTH-TOKEN: ${TOKEN}" \
  -H "Content-Type: application/json" \
  -X POST "https://<objectscale-host>:4443/object/bucket" \
  -d '{"name": "test-bucket", "namespace": "default"}'

# Expected: 403 Forbidden or permission denied
```

### Test 4: Test with Exporter
```bash
# Run exporter with new credentials
docker run --rm \
  -e ECSENV_USERNAME=prometheus-monitor \
  -e ECSENV_PASSWORD=YourPassword \
  prometheus-emcecs-exporter:latest &

# Wait a moment, then query metrics
curl "http://localhost:9438/query?target=<objectscale-host>"

# Expected: Prometheus metrics output
```

---

## Required Permissions for Monitoring

The monitoring user needs **read-only** access to these API endpoints:

### Critical (Required):
- ✅ `GET /login` - Authentication
- ✅ `GET /user/whoami` - Token validation
- ✅ `GET /dashboard/zones/localzone` - Cluster metrics
- ✅ `GET /dashboard/zones/localzone/replicationgroups` - Replication data
- ✅ `GET /vdc/nodes` - Node information
- ✅ Port 9021 `/?ping` - Active connections
- ✅ Port 9101 `/stats/dt/DTInitStat` - Node stats

### Optional (Enhanced Monitoring):
- 📊 `POST /object/billing/buckets/{namespace}/info` - Billing data
- 📊 `POST /object/billing/namespace/info` - Namespace billing
- 📊 `GET /vdc/nodes/{nodeId}` - Detailed node info
- 📊 `GET /object/capacity` - Capacity details

### Should NOT Have Access To:
- ❌ `POST /object/bucket` - Create buckets
- ❌ `DELETE /object/bucket` - Delete buckets
- ❌ `PUT /vdc/nodes` - Modify nodes
- ❌ `POST /vdc/users` - Create users
- ❌ Any write/delete operations

---

## Security Best Practices

### Password Requirements:
```
Minimum length: 8 characters (recommend 16+)
Must contain:
  - Uppercase letters (A-Z)
  - Lowercase letters (a-z)
  - Numbers (0-9)
  - Special characters (!@#$%^&*)

Example: ProM3theus!M0nit0r#2024
```

### Storage:
- **DO NOT** hardcode passwords in code or docker-compose files
- Use environment variables: `-e ECSENV_PASSWORD`
- Use Docker secrets in Swarm: `--secret ecs_password`
- Use Kubernetes secrets: `kubectl create secret`
- Use vault systems: HashiCorp Vault, AWS Secrets Manager

### Example with Docker Secrets:
```bash
# Create secret
echo "YourStrongPassword" | docker secret create ecs_password -

# Use in service
docker service create \
  --name ecs-exporter \
  --secret ecs_password \
  -e ECSENV_USERNAME=prometheus-monitor \
  -e ECSENV_PASSWORD_FILE=/run/secrets/ecs_password \
  prometheus-emcecs-exporter:latest
```

### Credential Rotation:
```bash
# Change password via API
curl -k -X PUT \
  -H "X-SDS-AUTH-TOKEN: ${ADMIN_TOKEN}" \
  -H "Content-Type: application/json" \
  "https://<objectscale-host>:4443/vdc/users/mgmt/prometheus-monitor" \
  -d '{
    "password": "NewStrongPassword456!"
  }'

# Update in exporter
docker restart ecs-exporter
```

### Network Security:
- **Firewall**: Only allow exporter host to access management API
- **TLS**: Always use HTTPS (port 4443)
- **Certificates**: Install proper CA certificates for production

---

## Troubleshooting

### Issue: "Authentication Failed"
```bash
# Check credentials
curl -k -v -u "prometheus-monitor:password" \
  "https://<objectscale-host>:4443/login" 2>&1 | grep "HTTP"

# Expected: HTTP/1.1 200 OK
# If 401: Wrong username/password
# If 403: User exists but may be locked
```

### Issue: "Permission Denied on Dashboard API"
```bash
# Verify user role
curl -k -H "X-SDS-AUTH-TOKEN: ${ADMIN_TOKEN}" \
  "https://<objectscale-host>:4443/vdc/users/mgmt/prometheus-monitor" \
  | jq '.isSystemMonitor'

# Expected: true
# If false: Re-assign System Monitor role
```

### Issue: "Cannot Access Nodes API"
```bash
# Check network connectivity
curl -k -v "https://<objectscale-host>:4443/vdc/nodes" \
  -H "X-SDS-AUTH-TOKEN: ${TOKEN}" 2>&1 | grep "HTTP"

# Ensure ports are open: 4443, 9021, 9101
```

### Issue: "Token Expired"
```bash
# Tokens typically expire after 8 hours
# The exporter handles token refresh automatically

# Manual token refresh:
NEW_TOKEN=$(curl -k -H "X-SDS-AUTH-TOKEN: ${OLD_TOKEN}" \
  "https://<objectscale-host>:4443/user/whoami" \
  | grep -oP 'X-SDS-AUTH-TOKEN: \K[^"]+')
```

---

## Verification Checklist

After creating the monitoring user, verify:

- [ ] User can authenticate (`/login` returns token)
- [ ] User can access dashboard API
- [ ] User can access VDC nodes API
- [ ] User **cannot** create/delete buckets
- [ ] User **cannot** modify node configuration
- [ ] Password meets security requirements
- [ ] Credentials stored securely (not in plaintext files)
- [ ] Exporter successfully scrapes metrics
- [ ] Prometheus successfully scrapes exporter

---

## Quick Setup Script

```bash
#!/bin/bash
# create-monitoring-user.sh

OBJECTSCALE_HOST="${1:-objectscale.example.com}"
ADMIN_USER="${2:-admin}"
ADMIN_PASS="${3}"
MONITOR_USER="prometheus-monitor"
MONITOR_PASS=$(openssl rand -base64 16)

echo "Creating monitoring user on ${OBJECTSCALE_HOST}..."

# Get admin token
ADMIN_TOKEN=$(curl -sk -u "${ADMIN_USER}:${ADMIN_PASS}" \
  "https://${OBJECTSCALE_HOST}:4443/login" \
  | grep -oP 'X-SDS-AUTH-TOKEN: \K[^"]+')

if [ -z "$ADMIN_TOKEN" ]; then
  echo "ERROR: Failed to authenticate as admin"
  exit 1
fi

# Create monitoring user
RESPONSE=$(curl -sk -X POST \
  -H "X-SDS-AUTH-TOKEN: ${ADMIN_TOKEN}" \
  -H "Content-Type: application/json" \
  "https://${OBJECTSCALE_HOST}:4443/vdc/users/mgmt" \
  -d "{
    \"user\": \"${MONITOR_USER}\",
    \"password\": \"${MONITOR_PASS}\",
    \"isSystemMonitor\": true
  }")

if echo "$RESPONSE" | grep -q "prometheus-monitor"; then
  echo "✓ User created successfully"
  echo ""
  echo "Credentials:"
  echo "Username: ${MONITOR_USER}"
  echo "Password: ${MONITOR_PASS}"
  echo ""
  echo "Save these credentials securely!"
  echo ""
  echo "Test with:"
  echo "docker run -p 9438:9438 \\"
  echo "  -e ECSENV_USERNAME=${MONITOR_USER} \\"
  echo "  -e ECSENV_PASSWORD='${MONITOR_PASS}' \\"
  echo "  prometheus-emcecs-exporter:latest"
else
  echo "ERROR: Failed to create user"
  echo "$RESPONSE"
  exit 1
fi
```

Usage:
```bash
chmod +x create-monitoring-user.sh
./create-monitoring-user.sh objectscale.example.com admin adminpassword
```

---

## References

- ObjectScale 4.1 API Documentation: `/ObjectScale_4.1_REST_API/`
- Management User API: `MgmtUserInfoService`
- Exporter Configuration: `README.md`

---

**Document Status**: Complete
**Next Review**: 2025-11-19
