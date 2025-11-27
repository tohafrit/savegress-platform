# Your First CDC Pipeline

This tutorial walks you through creating a complete CDC pipeline from PostgreSQL to a webhook endpoint.

## What You'll Build

```
PostgreSQL → Savegress Engine → Webhook Endpoint
```

**Time:** 15-20 minutes

## Prerequisites

- Docker installed
- PostgreSQL database (or use the included one)
- A webhook endpoint (or use webhook.site for testing)

## Step 1: Set Up the Environment

Create a project directory:

```bash
mkdir savegress-demo
cd savegress-demo
```

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  # PostgreSQL database
  postgres:
    image: postgres:16
    environment:
      POSTGRES_USER: demo
      POSTGRES_PASSWORD: demo123
      POSTGRES_DB: demo
    command:
      - postgres
      - -c
      - wal_level=logical
      - -c
      - max_replication_slots=10
      - -c
      - max_wal_senders=10
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql

  # Savegress Engine
  savegress:
    image: savegress/engine:latest
    depends_on:
      - postgres
    volumes:
      - ./config.yaml:/etc/savegress/config.yaml
    environment:
      - SAVEGRESS_LOG_LEVEL=info

volumes:
  postgres_data:
```

## Step 2: Initialize the Database

Create `init.sql`:

```sql
-- Create replication user
CREATE ROLE cdc_user WITH REPLICATION LOGIN PASSWORD 'cdc_password';

-- Create sample tables
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Grant permissions
GRANT SELECT ON ALL TABLES IN SCHEMA public TO cdc_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO cdc_user;

-- Create publication for CDC
CREATE PUBLICATION savegress_demo FOR TABLE users, orders;

-- Set full replica identity for complete before/after data
ALTER TABLE users REPLICA IDENTITY FULL;
ALTER TABLE orders REPLICA IDENTITY FULL;

-- Insert sample data
INSERT INTO users (name, email) VALUES
    ('Alice', 'alice@example.com'),
    ('Bob', 'bob@example.com'),
    ('Charlie', 'charlie@example.com');

INSERT INTO orders (user_id, amount, status) VALUES
    (1, 99.99, 'completed'),
    (2, 149.50, 'pending'),
    (1, 25.00, 'completed');
```

## Step 3: Configure Savegress

Create `config.yaml`:

```yaml
# Savegress Engine Configuration
source:
  type: postgres
  host: postgres
  port: 5432
  database: demo
  user: cdc_user
  password: cdc_password

  # Replication settings
  slot_name: savegress_demo_slot
  publication: savegress_demo
  create_slot: true

  # Tables to capture
  tables:
    - public.users
    - public.orders

# Output to console (for testing)
output:
  type: stdout
  format: json
  pretty: true

# Checkpointing
checkpoint:
  dir: /var/lib/savegress/checkpoints
  interval: 10s

# Logging
logging:
  level: info
  format: text
```

## Step 4: Start the Pipeline

```bash
# Start everything
docker-compose up -d

# Watch the logs
docker-compose logs -f savegress
```

You should see:

```
INFO  Starting Savegress Engine v1.0.0
INFO  Connecting to PostgreSQL postgres:5432/demo
INFO  Created replication slot: savegress_demo_slot
INFO  Subscribing to publication: savegress_demo
INFO  Tracking tables: public.users, public.orders
INFO  Pipeline started, waiting for changes...
```

## Step 5: Make Changes

Open another terminal and connect to PostgreSQL:

```bash
docker exec -it savegress-demo-postgres-1 psql -U demo -d demo
```

Make some changes:

```sql
-- Insert a new user
INSERT INTO users (name, email) VALUES ('Diana', 'diana@example.com');

-- Update a user
UPDATE users SET status = 'inactive' WHERE name = 'Bob';

-- Create an order
INSERT INTO orders (user_id, amount) VALUES (4, 75.00);

-- Delete a user (will fail due to FK, but try anyway)
DELETE FROM users WHERE name = 'Charlie';
```

## Step 6: See the Events

In the Savegress logs, you'll see events:

```json
{
  "id": "evt-001",
  "operation": "INSERT",
  "schema": "public",
  "table": "users",
  "timestamp": "2025-01-15T10:30:00Z",
  "after": {
    "id": 4,
    "name": "Diana",
    "email": "diana@example.com",
    "status": "active",
    "created_at": "2025-01-15T10:30:00"
  }
}
```

```json
{
  "id": "evt-002",
  "operation": "UPDATE",
  "schema": "public",
  "table": "users",
  "before": {
    "id": 2,
    "name": "Bob",
    "email": "bob@example.com",
    "status": "active"
  },
  "after": {
    "id": 2,
    "name": "Bob",
    "email": "bob@example.com",
    "status": "inactive"
  }
}
```

## Step 7: Add Webhook Output (Pro)

Now let's send events to a webhook. First, get a test endpoint from https://webhook.site.

Update `config.yaml`:

```yaml
source:
  type: postgres
  host: postgres
  port: 5432
  database: demo
  user: cdc_user
  password: cdc_password
  slot_name: savegress_demo_slot
  publication: savegress_demo
  tables:
    - public.users
    - public.orders

# Send to webhook (Pro feature)
output:
  type: webhook
  url: https://webhook.site/your-unique-url
  method: POST
  headers:
    Content-Type: application/json
    X-Source: savegress-demo

  # Batching (optional)
  batch_size: 10
  batch_timeout: 5s

  # Retry on failure
  retry:
    enabled: true
    max_attempts: 3
    delay: 1s

checkpoint:
  dir: /var/lib/savegress/checkpoints
  interval: 10s

logging:
  level: info
```

Restart Savegress:

```bash
docker-compose restart savegress
```

Make more changes in PostgreSQL and watch them appear at webhook.site!

## Step 8: Add Monitoring (Pro)

Enable Prometheus metrics:

```yaml
source:
  # ... same as before ...

output:
  # ... same as before ...

metrics:
  enabled: true
  address: :8080
  path: /metrics

  prometheus:
    enabled: true
    namespace: savegress_demo
```

Add to `docker-compose.yml`:

```yaml
services:
  savegress:
    # ... existing config ...
    ports:
      - "8080:8080"  # Expose metrics
```

Check metrics:

```bash
curl http://localhost:8080/metrics | grep savegress
```

## Step 9: Clean Up

```bash
# Stop and remove everything
docker-compose down -v

# Remove files
rm -rf savegress-demo
```

## What You Learned

1. **Database Setup:** Configure PostgreSQL for logical replication
2. **Savegress Config:** Create a pipeline configuration
3. **Event Capture:** See INSERT, UPDATE, DELETE events
4. **Webhook Output:** Send events to external systems
5. **Monitoring:** Enable Prometheus metrics

## Next Steps

- [Configuration Reference](../configuration/reference.md) - All configuration options
- [PostgreSQL Guide](../connectors/sources/postgresql.md) - Advanced PostgreSQL setup
- [Webhook Output](../connectors/sinks/webhook.md) - Webhook configuration
- [Compression](../features/compression.md) - Enable compression (Pro)
- [DLQ](../features/dlq.md) - Handle failed deliveries (Pro)

## Common Issues

### No Events Appearing

1. Check replication slot:
```sql
SELECT * FROM pg_replication_slots WHERE slot_name = 'savegress_demo_slot';
```

2. Check publication:
```sql
SELECT * FROM pg_publication_tables WHERE pubname = 'savegress_demo';
```

3. Check logs:
```bash
docker-compose logs savegress
```

### Connection Refused

1. Ensure PostgreSQL started first:
```bash
docker-compose up -d postgres
sleep 5
docker-compose up -d savegress
```

2. Check network:
```bash
docker network ls
docker network inspect savegress-demo_default
```

### Webhook Not Receiving

1. Check webhook URL is correct
2. Check Savegress logs for errors
3. Try stdout output first to verify events are captured
