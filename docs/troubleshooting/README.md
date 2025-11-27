# Troubleshooting Guide

Common issues and solutions for Savegress CDC.

## Quick Diagnostics

```bash
# Check engine status
savegress-engine --status

# Check broker status
savegress-broker --status

# View recent logs
journalctl -u savegress-engine --since "1 hour ago"

# Check metrics
curl -s http://localhost:8080/metrics | grep savegress

# Test database connection
savegress-engine --test-connection
```

---

## Connection Issues

### Cannot Connect to Database

**Symptoms:**
```
ERROR: failed to connect to database: connection refused
ERROR: dial tcp 192.168.1.100:5432: connect: connection refused
```

**Solutions:**

1. **Check network connectivity:**
```bash
# Test TCP connection
telnet db.example.com 5432
nc -zv db.example.com 5432

# Check firewall
sudo iptables -L -n | grep 5432
```

2. **Verify database is running:**
```bash
# PostgreSQL
sudo systemctl status postgresql
pg_isready -h localhost -p 5432

# MySQL
sudo systemctl status mysql
mysqladmin -h localhost -P 3306 ping
```

3. **Check authentication:**
```bash
# PostgreSQL
psql -h db.example.com -U cdc_user -d mydb

# MySQL
mysql -h db.example.com -u cdc_user -p mydb
```

4. **Verify configuration:**
```yaml
source:
  host: db.example.com  # Not localhost if in Docker
  port: 5432
  user: cdc_user
  password: correct_password
```

### SSL/TLS Connection Errors

**Symptoms:**
```
ERROR: SSL connection required but not configured
ERROR: certificate verify failed
```

**Solutions:**

1. **Enable TLS:**
```yaml
source:
  tls:
    enabled: true
    mode: verify-full
    ca_file: /path/to/ca.crt
```

2. **For development (skip verify):**
```yaml
source:
  tls:
    enabled: true
    mode: require
    skip_verify: true  # NOT for production
```

3. **Check certificate:**
```bash
openssl s_client -connect db.example.com:5432 -starttls postgres
```

### Replication Permission Denied

**Symptoms:**
```
ERROR: must be superuser or replication role to create replication slot
ERROR: permission denied for replication
```

**Solutions:**

PostgreSQL:
```sql
-- Grant replication permission
ALTER ROLE cdc_user WITH REPLICATION;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO cdc_user;

-- Verify
SELECT rolname, rolreplication FROM pg_roles WHERE rolname = 'cdc_user';
```

MySQL:
```sql
-- Grant permissions
GRANT SELECT, REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO 'cdc_user'@'%';
FLUSH PRIVILEGES;

-- Verify
SHOW GRANTS FOR 'cdc_user'@'%';
```

---

## No Events Captured

### PostgreSQL: No Events

**Check 1: WAL level**
```sql
SHOW wal_level;
-- Must be 'logical'
```

If not logical:
```bash
# Edit postgresql.conf
wal_level = logical
max_replication_slots = 10
max_wal_senders = 10

# Restart PostgreSQL (required)
sudo systemctl restart postgresql
```

**Check 2: Publication exists**
```sql
SELECT * FROM pg_publication WHERE pubname = 'savegress_pub';
SELECT * FROM pg_publication_tables WHERE pubname = 'savegress_pub';
```

If missing:
```sql
CREATE PUBLICATION savegress_pub FOR ALL TABLES;
-- Or specific tables
CREATE PUBLICATION savegress_pub FOR TABLE public.users, public.orders;
```

**Check 3: Replication slot**
```sql
SELECT * FROM pg_replication_slots WHERE slot_name = 'savegress_slot';
```

If missing or inactive:
```sql
-- Drop and recreate
SELECT pg_drop_replication_slot('savegress_slot');
SELECT pg_create_logical_replication_slot('savegress_slot', 'pgoutput');
```

**Check 4: Replica identity**
```sql
SELECT relname, relreplident FROM pg_class WHERE relname = 'users';
-- 'd' = default (PK only)
-- 'f' = full (all columns)
-- 'n' = nothing

-- For full before/after:
ALTER TABLE users REPLICA IDENTITY FULL;
```

### MySQL: No Events

**Check 1: Binary logging**
```sql
SHOW VARIABLES LIKE 'log_bin';
-- Must be ON

SHOW VARIABLES LIKE 'binlog_format';
-- Must be ROW

SHOW VARIABLES LIKE 'binlog_row_image';
-- Should be FULL
```

If not configured:
```ini
# my.cnf
[mysqld]
server-id = 1
log_bin = mysql-bin
binlog_format = ROW
binlog_row_image = FULL
```

**Check 2: Server ID**
```yaml
source:
  type: mysql
  server_id: 12345  # Must be unique
```

**Check 3: Position**
```sql
SHOW MASTER STATUS;
-- Note File and Position
```

### Common: Table Not in Config

**Check tables configuration:**
```yaml
source:
  tables:
    - public.users    # Schema-qualified
    - public.orders

  exclude_tables:
    - public.logs     # Make sure target isn't excluded
```

**Wildcard issues:**
```yaml
source:
  tables:
    - public.*        # All tables in public schema
    - "*.users"       # users table in all schemas
```

---

## Performance Issues

### High Latency

**Symptoms:**
- Event delay > 1 second
- Growing lag

**Diagnostics:**
```bash
# Check replication lag (PostgreSQL)
psql -c "SELECT slot_name, pg_size_pretty(pg_wal_lsn_diff(pg_current_wal_lsn(), restart_lsn)) AS lag FROM pg_replication_slots;"

# Check Savegress metrics
curl -s http://localhost:8080/metrics | grep savegress_lag
```

**Solutions:**

1. **Increase batch size:**
```yaml
batching:
  max_size: 500     # Larger batches
  max_wait: 50ms    # Shorter wait
```

2. **Enable compression:**
```yaml
compression:
  enabled: true
  algorithm: lz4    # Fast compression
```

3. **Check output bottleneck:**
```yaml
output:
  type: webhook
  timeout: 5s       # Reduce if target is slow
  retry:
    max_attempts: 3 # Reduce retries
```

4. **Increase parallelism:**
```yaml
parallel:
  table_parallelism: 8
  transaction_parallelism: 4
```

### High Memory Usage

**Symptoms:**
- OOM kills
- Memory > 2GB for simple workloads

**Solutions:**

1. **Reduce buffer size:**
```yaml
buffer:
  size: 4096        # Smaller buffer
  pool:
    max_size: 32    # Fewer pooled buffers
```

2. **Enable disk overflow:**
```yaml
buffer:
  overflow:
    enabled: true
    path: /var/lib/savegress/overflow
    max_size: 1GB
```

3. **Reduce batch size:**
```yaml
batching:
  max_size: 100     # Smaller batches use less memory
```

### High CPU Usage

**Symptoms:**
- CPU > 80%
- Throttling

**Solutions:**

1. **Reduce compression level:**
```yaml
compression:
  algorithm: lz4
  lz4:
    level: 1        # Fastest
```

2. **Reduce parallelism:**
```yaml
parallel:
  table_parallelism: 4   # Reduce from 8
```

3. **Filter unnecessary tables:**
```yaml
source:
  exclude_tables:
    - public.logs
    - public.sessions
```

---

## Output Issues

### Webhook Failures

**Symptoms:**
```
ERROR: webhook delivery failed: 503 Service Unavailable
ERROR: webhook delivery failed: timeout
```

**Solutions:**

1. **Increase timeout:**
```yaml
output:
  type: webhook
  timeout: 60s      # Increase for slow endpoints
```

2. **Configure retry:**
```yaml
output:
  retry:
    enabled: true
    max_attempts: 10
    initial_delay: 1s
    max_delay: 60s
```

3. **Enable DLQ for failures:**
```yaml
dlq:
  enabled: true
  max_retries: 5
```

4. **Check endpoint:**
```bash
curl -X POST https://your-webhook.com/events \
  -H "Content-Type: application/json" \
  -d '{"test": true}'
```

### Kafka Connection Issues

**Symptoms:**
```
ERROR: kafka: client has run out of available brokers
ERROR: kafka: dial tcp: connection refused
```

**Solutions:**

1. **Verify brokers:**
```yaml
output:
  type: kafka
  brokers:
    - kafka1:9092
    - kafka2:9092
    - kafka3:9092  # Multiple for failover
```

2. **Check SASL:**
```yaml
output:
  sasl:
    enabled: true
    mechanism: SCRAM-SHA-512
    username: ${KAFKA_USER}
    password: ${KAFKA_PASSWORD}
```

3. **Test connection:**
```bash
kafka-console-producer --broker-list kafka1:9092 --topic test
```

---

## License Issues

### License Expired

**Symptoms:**
```
ERROR: license expired on 2025-01-01
WARN: entering grace period (7 days remaining)
```

**Solutions:**

1. **Check license status:**
```bash
savegress-engine --license-info
```

2. **Renew license:**
   - Contact sales@savegress.io
   - Apply new license key

3. **Grace period:**
   - 7-day grace period after expiry
   - All features continue working
   - Renew before grace period ends

### Feature Not Licensed

**Symptoms:**
```
ERROR: feature 'compression' requires Pro license
ERROR: feature 'pitr' requires Enterprise license
```

**Solutions:**

1. **Check current tier:**
```bash
savegress-engine --license-info
# Tier: Community
# Features: postgresql, mysql, mariadb
```

2. **Upgrade license:**
   - Pro for compression, DLQ, Kafka
   - Enterprise for PITR, HA, Oracle

3. **Use alternative:**
```yaml
# Instead of compression (Pro)
batching:
  max_size: 1000  # Larger batches reduce overhead
```

### Limit Exceeded

**Symptoms:**
```
ERROR: source limit reached (1/1)
ERROR: table limit exceeded (10/10)
WARN: throughput throttled (1000/sec limit)
```

**Solutions:**

1. **Check usage:**
```bash
savegress-engine --license-info
# Usage:
#   Sources: 1/1 (100%)
#   Tables: 10/10 (100%)
```

2. **Optimize within limits:**
```yaml
source:
  tables:
    - public.critical_table1
    - public.critical_table2
  exclude_tables:
    - public.logs
    - public.sessions
```

3. **Upgrade for more:**
   - Pro: 10 sources, 100 tables, 50K/sec
   - Enterprise: Unlimited

---

## Storage Issues

### Disk Full

**Symptoms:**
```
ERROR: no space left on device
ERROR: write failed: disk quota exceeded
```

**Solutions:**

1. **Check disk usage:**
```bash
df -h /var/lib/savegress
du -sh /var/lib/savegress/*
```

2. **Reduce retention:**
```yaml
dlq:
  retention_days: 7    # Reduce from 14

checkpoint:
  interval: 30s        # Less frequent
```

3. **Clean old data:**
```bash
# Remove old checkpoints
find /var/lib/savegress/checkpoints -mtime +7 -delete

# Purge DLQ
savegress-cli dlq purge --older-than 7d
```

4. **Enable compression:**
```yaml
dlq:
  compression: true
```

### Checkpoint Corruption

**Symptoms:**
```
ERROR: failed to load checkpoint: invalid format
ERROR: checkpoint file corrupted
```

**Solutions:**

1. **Reset checkpoint:**
```bash
# Backup current
mv /var/lib/savegress/checkpoints /var/lib/savegress/checkpoints.bak

# Engine will start from latest position
savegress-engine --config config.yaml
```

2. **Start from specific position:**
```yaml
source:
  type: postgres
  # Start from specific LSN
  start_position: "0/1234ABCD"
```

---

## Diagnostic Commands

### Full Health Check

```bash
#!/bin/bash
echo "=== Savegress Health Check ==="

echo -e "\n--- Service Status ---"
systemctl status savegress-engine --no-pager
systemctl status savegress-broker --no-pager

echo -e "\n--- Resource Usage ---"
ps aux | grep savegress
df -h /var/lib/savegress

echo -e "\n--- Recent Errors ---"
journalctl -u savegress-engine --since "1 hour ago" -p err --no-pager

echo -e "\n--- Key Metrics ---"
curl -s http://localhost:8080/metrics | grep -E "savegress_(events|lag|errors)"

echo -e "\n--- Database Connection ---"
savegress-engine --test-connection

echo -e "\n--- License Status ---"
savegress-engine --license-info
```

### Collect Support Bundle

```bash
#!/bin/bash
BUNDLE_DIR="/tmp/savegress-support-$(date +%Y%m%d-%H%M%S)"
mkdir -p $BUNDLE_DIR

# Config (sanitized)
cp /etc/savegress/*.yaml $BUNDLE_DIR/
sed -i 's/password:.*/password: REDACTED/' $BUNDLE_DIR/*.yaml

# Logs
journalctl -u savegress-engine --since "24 hours ago" > $BUNDLE_DIR/engine.log
journalctl -u savegress-broker --since "24 hours ago" > $BUNDLE_DIR/broker.log

# Metrics
curl -s http://localhost:8080/metrics > $BUNDLE_DIR/metrics.txt

# Status
savegress-engine --status > $BUNDLE_DIR/status.txt
savegress-engine --license-info > $BUNDLE_DIR/license.txt

# System info
uname -a > $BUNDLE_DIR/system.txt
df -h >> $BUNDLE_DIR/system.txt
free -m >> $BUNDLE_DIR/system.txt

# Create archive
tar -czf ${BUNDLE_DIR}.tar.gz -C /tmp $(basename $BUNDLE_DIR)
echo "Support bundle: ${BUNDLE_DIR}.tar.gz"
```

---

## Getting Help

### Community Support

- **GitHub Issues:** https://github.com/savegress/savegress/issues
- **Discussions:** https://github.com/savegress/savegress/discussions

### Pro/Enterprise Support

- **Email:** support@savegress.io
- **Response time:** 24h (Pro), 4h (Enterprise)

### Before Contacting Support

1. Check this troubleshooting guide
2. Review logs for error messages
3. Collect support bundle
4. Include:
   - Savegress version
   - Database type and version
   - Configuration (sanitized)
   - Error messages
   - Steps to reproduce
