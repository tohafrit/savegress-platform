# Compression

Savegress Pro and Enterprise include hybrid compression that can reduce storage and bandwidth by 4-10x.

**License:** Pro+

## Overview

CDC events are often highly compressible because:
- Schema information is repetitive
- JSON structure is predictable
- Similar events occur in batches

Savegress compression addresses this with intelligent algorithm selection.

## Quick Start

```yaml
compression:
  enabled: true
  algorithm: hybrid  # Recommended
```

That's it! Hybrid mode automatically selects the best algorithm for each batch.

## Compression Algorithms

### Hybrid (Recommended)

Automatically selects the best algorithm based on data characteristics:

```yaml
compression:
  algorithm: hybrid
  hybrid:
    threshold: 4096       # Bytes
    small_algo: lz4       # Fast for small batches
    large_algo: zstd      # Better ratio for large batches
    large_level: 3
```

**How it works:**
- Data < 4KB → LZ4 (speed priority)
- Data ≥ 4KB → ZSTD (ratio priority)

### LZ4

Extremely fast compression with moderate ratio.

```yaml
compression:
  algorithm: lz4
  lz4:
    level: 3  # 1-12: higher = slower, better ratio
```

**Best for:**
- Ultra-low latency requirements (< 10ms)
- High throughput with CPU constraints
- Small messages

**Performance:**
- Compression: 400-500 MB/s
- Decompression: 1000+ MB/s
- Ratio: 2-3x

### ZSTD

Balanced compression with excellent ratio.

```yaml
compression:
  algorithm: zstd
  zstd:
    level: 3  # 1-22: higher = slower, better ratio
```

**Best for:**
- Bandwidth-constrained environments
- Large batch sizes
- Storage optimization

**Performance:**
- Compression: 200-400 MB/s
- Decompression: 500-800 MB/s
- Ratio: 4-10x

### SIMD Optimization (Enterprise)

Hardware-accelerated compression using CPU vector instructions.

```yaml
compression:
  algorithm: zstd
  simd:
    enabled: true
    instruction_set: auto  # auto, avx2, avx512, neon
```

**Supported instruction sets:**
- **AVX2:** Intel Haswell+, AMD Zen+
- **AVX-512:** Intel Skylake-X+, AMD Zen 4+
- **NEON:** ARM64 (Apple Silicon, AWS Graviton)

**Performance boost:**
- 2-3x faster compression
- Near-native decompression speed

## Configuration Reference

### Full Configuration

```yaml
compression:
  enabled: true
  algorithm: hybrid

  # Minimum size to compress (skip small messages)
  min_size: 256

  # LZ4 settings
  lz4:
    level: 3              # 1-12

  # ZSTD settings
  zstd:
    level: 3              # 1-22
    window_log: 0         # 0 = auto

  # Hybrid settings
  hybrid:
    threshold: 4096       # Bytes
    small_algo: lz4
    small_level: 3
    large_algo: zstd
    large_level: 5

  # SIMD settings (Enterprise)
  simd:
    enabled: true
    instruction_set: auto

  # Compression pool
  pool:
    enabled: true
    size: 4               # Number of compressors
```

### Level Guidelines

| Use Case | LZ4 Level | ZSTD Level |
|----------|-----------|------------|
| Ultra-low latency | 1 | 1 |
| Low latency | 3 | 3 |
| Balanced | 5 | 5 |
| High compression | 9 | 9 |
| Maximum compression | 12 | 15-19 |

## Performance Tuning

### Latency-Optimized

```yaml
# Target: < 10ms added latency
compression:
  algorithm: lz4
  lz4:
    level: 1
  min_size: 1024  # Skip small messages
```

### Throughput-Optimized

```yaml
# Target: Maximum events/sec
compression:
  algorithm: hybrid
  hybrid:
    threshold: 2048
    small_algo: lz4
    small_level: 1
    large_algo: zstd
    large_level: 3
  pool:
    size: 8  # Match CPU cores
```

### Storage-Optimized

```yaml
# Target: Minimum storage/bandwidth
compression:
  algorithm: zstd
  zstd:
    level: 9
  simd:
    enabled: true  # Enterprise
```

### Batch Processing

```yaml
# Target: Non-real-time, maximum compression
compression:
  algorithm: zstd
  zstd:
    level: 15
  min_size: 0  # Compress everything
```

## Compression Ratio by Data Type

Typical compression ratios for CDC events:

| Data Type | LZ4 Ratio | ZSTD Ratio |
|-----------|-----------|------------|
| JSON (structured) | 3-4x | 6-10x |
| Text-heavy | 2-3x | 4-6x |
| Numeric-heavy | 2-2.5x | 3-4x |
| Binary/BLOB | 1.1-1.5x | 1.2-2x |
| Already compressed | 1.0x | 1.0x |

## Metrics

### Prometheus Metrics

```prometheus
# Compression ratio
savegress_compression_ratio{algorithm="zstd"} 5.2

# Bytes before compression
savegress_compression_bytes_in_total{algorithm="zstd"} 1073741824

# Bytes after compression
savegress_compression_bytes_out_total{algorithm="zstd"} 206612889

# Compression time
savegress_compression_duration_seconds_bucket{algorithm="zstd",le="0.001"} 9500
savegress_compression_duration_seconds_bucket{algorithm="zstd",le="0.01"} 10000

# Compression pool utilization
savegress_compression_pool_active{} 4
savegress_compression_pool_idle{} 2
```

### Logging

```json
{
  "level": "debug",
  "msg": "batch compressed",
  "algorithm": "zstd",
  "original_size": 102400,
  "compressed_size": 19692,
  "ratio": 5.2,
  "duration_ms": 2.3
}
```

## Cost Savings Calculator

### Example: 1TB/day throughput

**Without compression:**
- Storage: 1 TB/day × $0.023/GB = $23/day = **$690/month**
- Bandwidth: 1 TB/day × $0.09/GB = $90/day = **$2,700/month**

**With ZSTD (5x ratio):**
- Storage: 200 GB/day × $0.023/GB = $4.60/day = **$138/month** (-80%)
- Bandwidth: 200 GB/day × $0.09/GB = $18/day = **$540/month** (-80%)

**Monthly savings: $2,712**

### Calculator

```
Original Size (GB/day) × 30 days × (Storage $/GB + Bandwidth $/GB) × (1 - 1/Ratio)
```

## Integration with Other Features

### Compression + Batching

Larger batches compress better:

```yaml
batching:
  max_size: 500      # Larger batches
  max_wait: 100ms

compression:
  enabled: true
  algorithm: hybrid
  min_size: 0        # Compress all batches
```

### Compression + DLQ

DLQ stores compressed messages to save space:

```yaml
compression:
  enabled: true

dlq:
  enabled: true
  compression: true  # Compress DLQ entries
```

### Compression + Cloud Storage

Compressed before upload to reduce transfer costs:

```yaml
compression:
  enabled: true
  algorithm: zstd
  zstd:
    level: 9         # Maximum compression for storage

storage:
  backend: s3
  # Data is compressed before upload
```

## Troubleshooting

### Low Compression Ratio

**Symptoms:** Ratio < 2x

**Causes:**
1. Already compressed data (images, PDFs)
2. High-entropy data (encrypted, random)
3. Very small messages

**Solutions:**
```yaml
compression:
  # Skip small messages
  min_size: 512

  # Skip specific tables with binary data
  exclude_tables:
    - public.attachments
    - public.files
```

### High CPU Usage

**Symptoms:** CPU > 80% during compression

**Solutions:**

1. Reduce compression level:
```yaml
compression:
  zstd:
    level: 1  # Fastest
```

2. Use LZ4:
```yaml
compression:
  algorithm: lz4
```

3. Increase pool size:
```yaml
compression:
  pool:
    size: 8  # More parallel compressors
```

### Increased Latency

**Symptoms:** Event latency increased after enabling compression

**Solutions:**

1. Use hybrid for mixed workloads:
```yaml
compression:
  algorithm: hybrid
  hybrid:
    threshold: 2048  # Lower threshold
```

2. Skip compression for real-time:
```yaml
compression:
  min_size: 10240  # Only compress large batches
```

## Best Practices

1. **Start with hybrid:** Auto-selects the best algorithm

2. **Enable SIMD when available:** Free 2-3x speedup on Enterprise

3. **Tune based on metrics:** Monitor `compression_ratio` and `compression_duration`

4. **Consider batch size:** Larger batches = better compression

5. **Exclude binary data:** Don't compress already-compressed data

6. **Test with production data:** Ratios vary by data type

## See Also

- [Batching](batching.md) - Batch configuration
- [Performance Tuning](../configuration/optimization.md) - Overall optimization
- [Metrics](../api/metrics.md) - Monitoring compression
