# Quick Start

Get Savegress running in 5 minutes.

## Prerequisites

- Docker and Docker Compose
- A PostgreSQL, MySQL, or MariaDB database

## Step 1: Download

```bash
# Clone the repository
git clone https://github.com/savegress/savegress.git
cd savegress

# Or download the binary directly
curl -LO https://github.com/savegress/savegress/releases/latest/download/savegress-linux-amd64
chmod +x savegress-linux-amd64
```

## Step 2: Configure Your Database

### PostgreSQL

```sql
-- Enable logical replication in postgresql.conf
-- wal_level = logical

-- Create replication user
CREATE ROLE cdc_user WITH REPLICATION LOGIN PASSWORD 'cdc_password';
GRANT SELECT ON ALL TABLES IN SCHEMA public TO cdc_user;

-- Create publication
CREATE PUBLICATION savegress_pub FOR ALL TABLES;
```

### MySQL

```sql
-- Enable binlog in my.cnf
-- server-id = 1
-- log_bin = mysql-bin
-- binlog_format = ROW
-- binlog_row_image = FULL

-- Create replication user
CREATE USER 'cdc_user'@'%' IDENTIFIED BY 'cdc_password';
GRANT SELECT, REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO 'cdc_user'@'%';
```

## Step 3: Create Configuration

```yaml
# config.yaml
source:
  type: postgres  # or mysql, mariadb
  host: localhost
  port: 5432
  database: mydb
  user: cdc_user
  password: cdc_password
  slot_name: savegress_slot
  publication: savegress_pub
  tables:
    - public.users
    - public.orders

output:
  type: stdout  # Print events to console
  # For production, use webhook, kafka, or grpc
```

## Step 4: Run Savegress

### Using Docker

```bash
docker run -d \
  --name savegress \
  -v $(pwd)/config.yaml:/etc/savegress/config.yaml \
  savegress/engine:latest
```

### Using Binary

```bash
./savegress-engine --config config.yaml
```

### Using Docker Compose

```yaml
# docker-compose.yml
version: '3.8'
services:
  savegress:
    image: savegress/engine:latest
    volumes:
      - ./config.yaml:/etc/savegress/config.yaml
    environment:
      - SAVEGRESS_LOG_LEVEL=info
```

```bash
docker-compose up -d
```

## Step 5: Make a Change

```sql
-- Insert a record
INSERT INTO users (name, email) VALUES ('John Doe', 'john@example.com');

-- Update a record
UPDATE users SET email = 'john.doe@example.com' WHERE name = 'John Doe';

-- Delete a record
DELETE FROM users WHERE name = 'John Doe';
```

## Step 6: See the Events

```bash
# View logs
docker logs -f savegress
```

You'll see events like:

```json
{
  "id": "evt-001",
  "operation": "INSERT",
  "table": "users",
  "timestamp": "2025-01-15T10:30:00Z",
  "after": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com"
  }
}
```

## Quick Configuration Examples

### Output to Webhook (Pro)

```yaml
output:
  type: webhook
  url: https://your-api.com/events
  headers:
    Authorization: "Bearer ${WEBHOOK_TOKEN}"
  batch_size: 100
  retry:
    max_attempts: 3
    delay: 1s
```

### Output to Kafka (Pro)

```yaml
output:
  type: kafka
  brokers:
    - kafka1:9092
    - kafka2:9092
  topic: cdc-events
  compression: snappy
```

### Multiple Tables with Filtering

```yaml
source:
  type: postgres
  # ... connection details ...

  tables:
    - public.users
    - public.orders
    - public.products

  exclude_tables:
    - public.sessions
    - public.logs

  # Only capture specific columns
  columns:
    public.users:
      - id
      - name
      - email
      # Excludes: password_hash, internal_notes
```

### With Compression (Pro)

```yaml
compression:
  enabled: true
  algorithm: hybrid  # auto-selects best algorithm
  # Saves 4-10x storage/bandwidth
```

## What's Next?

- [Installation Guide](installation.md) - Production deployment
- [First Pipeline](first-pipeline.md) - Step-by-step tutorial
- [Configuration Reference](../configuration/reference.md) - All options
- [PostgreSQL Setup](../connectors/sources/postgresql.md) - Detailed database setup

## Common Issues

### Connection Refused

```bash
# Check database is accessible
psql -h localhost -U cdc_user -d mydb

# Check firewall
telnet localhost 5432
```

### Permission Denied

```sql
-- PostgreSQL: Grant replication permission
ALTER ROLE cdc_user WITH REPLICATION;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO cdc_user;

-- MySQL: Grant replication permission
GRANT REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO 'cdc_user'@'%';
```

### No Events Captured

```sql
-- PostgreSQL: Check publication exists
SELECT * FROM pg_publication WHERE pubname = 'savegress_pub';

-- Check tables are in publication
SELECT * FROM pg_publication_tables WHERE pubname = 'savegress_pub';
```

See [Troubleshooting](../troubleshooting/README.md) for more solutions.
