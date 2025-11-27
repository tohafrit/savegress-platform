# Event Format

CDC events emitted by Savegress follow a consistent JSON schema.

## Event Structure

```json
{
  "id": "evt-abc123",
  "source": "postgres",
  "schema": "public",
  "table": "users",
  "operation": "UPDATE",
  "timestamp": "2025-01-15T10:30:00.123456Z",
  "transaction_id": "tx-12345",
  "position": {
    "lsn": "0/1234ABCD",
    "sequence": 42
  },
  "before": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com"
  },
  "after": {
    "id": 1,
    "name": "John Doe",
    "email": "john.doe@example.com"
  },
  "metadata": {
    "connector_version": "1.0.0",
    "database": "production",
    "server": "db1.example.com"
  }
}
```

## Field Reference

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique event identifier |
| `source` | string | Database type (postgres, mysql, etc.) |
| `schema` | string | Database schema name |
| `table` | string | Table name |
| `operation` | string | Operation type (INSERT, UPDATE, DELETE, DDL) |
| `timestamp` | string | ISO 8601 timestamp (UTC) |
| `transaction_id` | string | Database transaction ID |
| `position` | object | Position information for replay |
| `before` | object | Row state before change (UPDATE, DELETE) |
| `after` | object | Row state after change (INSERT, UPDATE) |
| `metadata` | object | Additional context |

## Operations

### INSERT

```json
{
  "operation": "INSERT",
  "before": null,
  "after": {
    "id": 1,
    "name": "New User",
    "email": "new@example.com",
    "created_at": "2025-01-15T10:30:00Z"
  }
}
```

### UPDATE

```json
{
  "operation": "UPDATE",
  "before": {
    "id": 1,
    "name": "Old Name",
    "email": "old@example.com"
  },
  "after": {
    "id": 1,
    "name": "New Name",
    "email": "new@example.com"
  }
}
```

### DELETE

```json
{
  "operation": "DELETE",
  "before": {
    "id": 1,
    "name": "Deleted User",
    "email": "deleted@example.com"
  },
  "after": null
}
```

### DDL (Schema Change)

```json
{
  "operation": "DDL",
  "schema": "public",
  "table": "users",
  "ddl_type": "ALTER_TABLE",
  "ddl_command": "ALTER TABLE users ADD COLUMN phone VARCHAR(20)",
  "before": null,
  "after": null
}
```

## Transaction Boundaries

Events include transaction markers when `preserve_transactions: true`:

```json
[
  {"operation": "BEGIN", "transaction_id": "tx-100", "timestamp": "..."},
  {"operation": "INSERT", "table": "orders", "transaction_id": "tx-100", ...},
  {"operation": "UPDATE", "table": "inventory", "transaction_id": "tx-100", ...},
  {"operation": "COMMIT", "transaction_id": "tx-100", "timestamp": "..."}
]
```

## Data Type Mappings

### PostgreSQL

| PostgreSQL | JSON | Example |
|------------|------|---------|
| integer, bigint | number | `42` |
| numeric, decimal | string | `"123.456"` |
| boolean | boolean | `true` |
| text, varchar | string | `"hello"` |
| timestamp | string | `"2025-01-15T10:30:00Z"` |
| timestamptz | string | `"2025-01-15T10:30:00+00:00"` |
| date | string | `"2025-01-15"` |
| time | string | `"10:30:00"` |
| uuid | string | `"550e8400-e29b-41d4-a716-446655440000"` |
| json, jsonb | object/array | `{"key": "value"}` |
| bytea | string (base64) | `"SGVsbG8gV29ybGQ="` |
| array | array | `[1, 2, 3]` |
| inet | string | `"192.168.1.1"` |
| point | object | `{"x": 1.0, "y": 2.0}` |

### MySQL

| MySQL | JSON | Example |
|-------|------|---------|
| INT, BIGINT | number | `42` |
| DECIMAL | string | `"123.456"` |
| TINYINT(1) | boolean | `true` |
| VARCHAR, TEXT | string | `"hello"` |
| DATETIME | string | `"2025-01-15 10:30:00"` |
| TIMESTAMP | string | `"2025-01-15T10:30:00Z"` |
| DATE | string | `"2025-01-15"` |
| TIME | string | `"10:30:00"` |
| JSON | object/array | `{"key": "value"}` |
| BLOB | string (base64) | `"SGVsbG8gV29ybGQ="` |
| ENUM | string | `"active"` |
| SET | array | `["a", "b"]` |

## Position Information

Position enables exactly-once processing and replay:

### PostgreSQL

```json
{
  "position": {
    "lsn": "0/1234ABCD",
    "sequence": 42
  }
}
```

### MySQL

```json
{
  "position": {
    "file": "mysql-bin.000001",
    "position": 12345,
    "gtid": "3E11FA47-71CA-11E1-9E33-C80AA9429562:1-5"
  }
}
```

### MongoDB

```json
{
  "position": {
    "resume_token": "82...",
    "cluster_time": {"t": 1234567890, "i": 1}
  }
}
```

## Metadata

```json
{
  "metadata": {
    "connector_version": "1.0.0",
    "database": "production",
    "server": "db1.example.com",
    "server_id": 12345,
    "pipeline_id": "pipe-abc123",
    "captured_at": "2025-01-15T10:30:00.123456Z"
  }
}
```

## Batch Format

When events are batched:

```json
{
  "batch_id": "batch-abc123",
  "batch_size": 100,
  "batch_timestamp": "2025-01-15T10:30:00Z",
  "events": [
    { "id": "evt-001", "operation": "INSERT", ... },
    { "id": "evt-002", "operation": "UPDATE", ... },
    { "id": "evt-003", "operation": "DELETE", ... }
  ]
}
```

## Webhook Delivery

When using webhook output:

```http
POST /events HTTP/1.1
Host: your-api.com
Content-Type: application/json
X-Savegress-Event-ID: evt-abc123
X-Savegress-Signature: sha256=abc123...
X-Savegress-Timestamp: 1705312200

{
  "id": "evt-abc123",
  "operation": "INSERT",
  ...
}
```

### Signature Verification

```python
import hmac
import hashlib

def verify_signature(payload, signature, secret):
    expected = hmac.new(
        secret.encode(),
        payload,
        hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(f"sha256={expected}", signature)
```

## Kafka Message Format

```
Key: users:1  (table:primary_key)
Value: {"id": "evt-abc123", ...}
Headers:
  - savegress-event-id: evt-abc123
  - savegress-operation: INSERT
  - savegress-table: users
  - savegress-timestamp: 1705312200000
```

## gRPC Message

```protobuf
message CDCEvent {
  string id = 1;
  string source = 2;
  string schema = 3;
  string table = 4;
  Operation operation = 5;
  google.protobuf.Timestamp timestamp = 6;
  string transaction_id = 7;
  Position position = 8;
  google.protobuf.Struct before = 9;
  google.protobuf.Struct after = 10;
  map<string, string> metadata = 11;
}

enum Operation {
  OPERATION_UNSPECIFIED = 0;
  INSERT = 1;
  UPDATE = 2;
  DELETE = 3;
  DDL = 4;
  BEGIN = 5;
  COMMIT = 6;
}
```
