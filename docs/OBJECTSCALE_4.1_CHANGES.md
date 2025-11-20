# ObjectScale 4.1 Changes

**Date**: 2024-11-19  
**Summary**: Updates to support Dell ObjectScale 4.1 API changes

---

## Key Changes

### Port 9101 DT Stats API Removed

**ObjectScale 4.1 no longer provides the DT (Data Table) statistics API on port 9101**. This API was present in ECS 3.6 but has been replaced by enhanced Dashboard API endpoints.

### New Dashboard API for Node Statistics

Node-level statistics are now retrieved via:

**Endpoint**: `GET /dashboard/nodes/{node-id}`  
**Port**: 4443 (Management API)  
**Authentication**: X-SDS-AUTH-TOKEN (same as before)

---

## New Metrics Available

The exporter now collects node-level metrics from the Dashboard API:

### Disk Metrics
- `emcecs_node_disks_total` - Total number of disks on node
- `emcecs_node_disks_good` - Number of good disks
- `emcecs_node_disks_bad` - Number of bad disks

### Storage Metrics
- `emcecs_node_disk_space_total_bytes` - Total disk space in bytes
- `emcecs_node_disk_space_free_bytes` - Free disk space in bytes
- `emcecs_node_disk_space_allocated_bytes` - Allocated disk space in bytes

### Connection Metrics
- `emcecs_node_active_connections` - Active S3/Swift connections (still from port 9021)

### API Limitations

**Note**: The ObjectScale 4.1 Dashboard API (`/dashboard/nodes/{node-id}`) does **not** provide the following metrics:
- ❌ CPU utilization metrics
- ❌ Memory utilization metrics
- ❌ Network bandwidth/utilization metrics
- ❌ Transaction latency/bandwidth metrics

These fields simply do not exist in the API response. Only disk and storage-related metrics are available from this endpoint.

---

## Removed Metrics

The following metrics are NO LONGER available (DT Stats API removed):

- ❌ `emcecs_node_dtTotal` - Total DT count
- ❌ `emcecs_node_dtUnready` - Unready DT count
- ❌ `emcecs_node_dtUnknown` - Unknown DT count

These metrics were specific to the internal DT (Data Table) storage structure and are not exposed in ObjectScale 4.1's Dashboard API.

---

## Migration Impact

### Breaking Changes
- **Metric names changed**: All node metrics renamed for clarity and consistency
- **Old DT metrics removed**: If you have dashboards using `dtTotal`, `dtUnready`, `dtUnknown`, these need updating

### Grafana Dashboard Updates Required

If you have existing Grafana dashboards, update queries:

**Old Query (ECS 3.6)**:
```promql
emcecs_node_dtTotal{node="<node-ip>"}
```

**New Query (ObjectScale 4.1)**:
```promql
emcecs_node_disks_total{node="<node-ip>"}
```

### Port Configuration

**Before (ECS 3.6)**:
- Port 4443 (Management API)
- Port 9021 (Object API)
- Port 9101 (DT Stats) ⚠️

**After (ObjectScale 4.1)**:
- Port 4443 (Management API + Node Stats)
- Port 9021 (Object API)
- ~~Port 9101~~ ❌ No longer required

---

## Compatibility

- ✅ **ObjectScale 4.1** - Fully supported
- ✅ **ECS 3.9** - Fully supported (uses same API as ObjectScale 4.1)
- ⚠️ **ECS 3.6** - May have limited node metrics (old DT stats may not work)

**Recommendation**: This version of the exporter is optimized for ObjectScale 4.1 / ECS 3.9+. For older ECS versions, use an earlier release of the exporter.

---

## Testing

After deploying the updated exporter:

1. **Verify metrics collection**:
   ```bash
   curl http://localhost:9438/query?target=<objectscale-host> | grep emcecs_node
   ```

2. **Check for new metrics**:
   ```bash
   curl http://localhost:9438/query?target=<objectscale-host> | grep cpu_utilization
   ```

3. **Enable debug mode** to troubleshoot:
   ```bash
   docker run -e ECSENV_DEBUG=true -e ECSENV_USERNAME=... prometheus-emcecs-exporter:latest
   ```

---

## Benefits

The Dashboard API provides:

1. **Working Node Metrics**: Disk and storage metrics now functional (were broken with port 9101 removal)
2. **Better Performance**: Single API call per node for disk/storage data
3. **Simplified Architecture**: No need for port 9101 access
4. **ObjectScale 4.1 Compatible**: Uses the supported API for the current version

---

## References

- API Compatibility Analysis: `/docs/API_COMPATIBILITY_ANALYSIS.md`
- Enhancement Opportunities: `/docs/ENHANCEMENT_OPPORTUNITIES.md`
- Read-Only User Guide: `/docs/CREATING_READONLY_USER.md`

---

**Document Version**: 1.0  
**Last Updated**: 2024-11-19
