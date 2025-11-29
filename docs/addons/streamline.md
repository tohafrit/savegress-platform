# StreamLine — E-commerce & Retail Add-on

> "Sync your commerce, everywhere"

## Overview

StreamLine — решение для omnichannel ритейла, которое использует Savegress CDC для синхронизации inventory, заказов и цен между множеством каналов продаж в реальном времени.

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          STREAMLINE                                      │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                 │
│   │   Shopify    │  │   Amazon     │  │  WooCommerce │                 │
│   └──────┬───────┘  └──────┬───────┘  └──────┬───────┘                 │
│          │                 │                 │                          │
│          └────────────────┬┴─────────────────┘                          │
│                           │                                              │
│                           ▼                                              │
│          ┌────────────────────────────────────┐                         │
│          │         StreamLine Hub             │                         │
│          │  ┌──────────┐  ┌──────────┐       │                         │
│          │  │ Inventory│  │  Order   │       │                         │
│          │  │  Engine  │  │  Router  │       │                         │
│          │  └──────────┘  └──────────┘       │                         │
│          │  ┌──────────┐  ┌──────────┐       │                         │
│          │  │  Price   │  │  Stock   │       │                         │
│          │  │  Sync    │  │  Alerts  │       │                         │
│          │  └──────────┘  └──────────┘       │                         │
│          └────────────────────────────────────┘                         │
│                           │                                              │
│                           ▼                                              │
│          ┌────────────────────────────────────┐                         │
│          │        Savegress CDC               │                         │
│          │   (Master Database Sync)           │                         │
│          └────────────────────────────────────┘                         │
│                           │                                              │
│          ┌────────────────┼────────────────┐                            │
│          ▼                ▼                ▼                            │
│   ┌──────────────┐ ┌──────────────┐ ┌──────────────┐                   │
│   │     ERP      │ │   Warehouse  │ │   Analytics  │                   │
│   │   (1C/SAP)   │ │    (WMS)     │ │    (BI)      │                   │
│   └──────────────┘ └──────────────┘ └──────────────┘                   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Target Audience

| Segment | Size | Use Case | Pain Point |
|---------|------|----------|------------|
| **D2C Brands** | 10K-100K SKUs | Multi-channel selling | Overselling |
| **Marketplaces** | 100K+ SKUs | Seller inventory sync | Stock accuracy |
| **Retail Chains** | 50+ stores | Store-to-web sync | Inventory visibility |
| **Distributors** | B2B + B2C | Channel-specific pricing | Price consistency |
| **Dropshippers** | Variable | Supplier sync | Stock availability |

---

## Core Features

### 1. Inventory Sync Engine

Real-time синхронизация inventory между всеми каналами.

```
┌─────────────────────────────────────────────────────────────────┐
│  Inventory Sync Engine                                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Product: SKU-12345 "Wireless Headphones"                       │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    Master Stock: 150                     │   │
│  └─────────────────────────────────────────────────────────┘   │
│                              │                                   │
│        ┌─────────────────────┼─────────────────────┐            │
│        │                     │                     │            │
│        ▼                     ▼                     ▼            │
│  ┌───────────┐        ┌───────────┐        ┌───────────┐       │
│  │  Shopify  │        │  Amazon   │        │   eBay    │       │
│  │           │        │           │        │           │       │
│  │  Alloc:50 │        │  Alloc:70 │        │  Alloc:30 │       │
│  │  Sold: 12 │        │  Sold: 25 │        │  Sold: 8  │       │
│  │  Avail:38 │        │  Avail:45 │        │  Avail:22 │       │
│  └───────────┘        └───────────┘        └───────────┘       │
│                                                                  │
│  Sync Status: ✓ All channels synchronized                       │
│  Last sync: 2 seconds ago                                        │
│  Conflicts resolved: 0                                           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

**Allocation Strategies:**
```yaml
inventory_allocation:
  strategy: weighted  # fixed, weighted, priority, dynamic

  channels:
    shopify:
      weight: 40%
      min_stock: 5
      buffer: 10%  # Safety buffer

    amazon:
      weight: 45%
      min_stock: 10
      buffer: 15%

    ebay:
      weight: 15%
      min_stock: 2
      buffer: 5%

  rules:
    - name: "Holiday boost"
      condition: "date between Dec 1 and Dec 25"
      action: "increase shopify weight by 20%"

    - name: "Low stock protection"
      condition: "total_stock < 20"
      action: "disable ebay listing"
```

---

### 2. Order Router

Умная маршрутизация заказов между fulfillment центрами.

```
┌─────────────────────────────────────────────────────────────────┐
│  Order Router                                                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  New Order: #ORD-78945                                          │
│  Channel: Amazon │ Items: 3 │ Ship to: Los Angeles, CA          │
│                                                                  │
│  Routing Decision:                                               │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                                                           │  │
│  │  Option 1: LA Warehouse (recommended)                     │  │
│  │  ├─ Distance: 12 miles                                    │  │
│  │  ├─ Stock: ✓ All items available                         │  │
│  │  ├─ Ship time: Same day                                   │  │
│  │  └─ Cost: $4.50                                           │  │
│  │                                                           │  │
│  │  Option 2: Phoenix Warehouse                              │  │
│  │  ├─ Distance: 370 miles                                   │  │
│  │  ├─ Stock: ✓ All items available                         │  │
│  │  ├─ Ship time: 2 days                                     │  │
│  │  └─ Cost: $8.20                                           │  │
│  │                                                           │  │
│  │  Option 3: Split shipment                                 │  │
│  │  ├─ Items 1,2: LA Warehouse                               │  │
│  │  ├─ Item 3: Chicago Warehouse                             │  │
│  │  └─ Cost: $12.80 (not recommended)                        │  │
│  │                                                           │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  [Auto-route to LA Warehouse] [Manual override] [Hold]          │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

**Routing Rules:**
```yaml
order_routing:
  optimization: cost  # cost, speed, stock_balance

  warehouses:
    - id: la_warehouse
      location: "Los Angeles, CA"
      priority: 1
      capabilities: [standard, express, fragile]

    - id: phoenix_warehouse
      location: "Phoenix, AZ"
      priority: 2
      capabilities: [standard, bulky]

    - id: chicago_warehouse
      location: "Chicago, IL"
      priority: 2
      capabilities: [standard, express, cold_chain]

  rules:
    - name: "Same-day eligible"
      condition: "shipping_method == 'express' AND distance < 50mi"
      action: "route to nearest with stock"

    - name: "Fragile items"
      condition: "any item has tag 'fragile'"
      action: "route to warehouse with 'fragile' capability"

    - name: "Avoid split"
      condition: "split_cost > single_warehouse_cost * 1.5"
      action: "route all to single warehouse"
```

---

### 3. Price Sync

Синхронизация цен с учётом специфики каждого канала.

```
┌─────────────────────────────────────────────────────────────────┐
│  Price Management                                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Product: SKU-12345 "Wireless Headphones"                       │
│  Base Price: $99.99                                              │
│                                                                  │
│  Channel Pricing:                                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ Channel      │ Fees    │ Margin  │ Final Price │ Status  │  │
│  │──────────────│─────────│─────────│─────────────│─────────│  │
│  │ Website      │ 3%      │ 40%     │ $99.99      │ ✓ Live  │  │
│  │ Shopify      │ 2.9%    │ 38%     │ $99.99      │ ✓ Live  │  │
│  │ Amazon       │ 15%     │ 25%     │ $109.99     │ ✓ Live  │  │
│  │ eBay         │ 13%     │ 27%     │ $104.99     │ ✓ Live  │  │
│  │ Walmart      │ 12%     │ 28%     │ $102.99     │ Pending │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  Price Rules Active:                                             │
│  • Amazon MAP compliance: min $99.99                             │
│  • eBay competitive: match lowest + $2                           │
│  • Weekend sale: -10% on website (Sat-Sun)                       │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

**Price Rules:**
```yaml
pricing_rules:
  global:
    min_margin: 20%
    round_to: 0.99  # $X.99 pricing

  channels:
    amazon:
      fee_percent: 15
      include_fba: true
      map_enforcement: true

    ebay:
      fee_percent: 13
      competitive_pricing:
        enabled: true
        strategy: match_lowest_plus
        adjustment: 2.00

    shopify:
      fee_percent: 2.9
      allow_discounts: true

  promotions:
    - name: "Weekend Sale"
      channels: [shopify, website]
      schedule: "SAT-SUN"
      discount: 10%

    - name: "Prime Day Match"
      channels: [shopify]
      trigger: "amazon_price_drop > 15%"
      action: "match amazon price"
```

---

### 4. Stock Alerts

Умные уведомления о состоянии inventory.

```
┌─────────────────────────────────────────────────────────────────┐
│  Stock Alerts                                        [Settings] │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  🔴 Critical (3)                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ SKU-789 "USB-C Cable" - OUT OF STOCK                     │  │
│  │ All channels affected │ Lost sales: ~$450/day            │  │
│  │ [Reorder] [Disable listings] [Find supplier]             │  │
│  ├──────────────────────────────────────────────────────────┤  │
│  │ SKU-456 "Phone Case" - OVERSOLD by 5 units               │  │
│  │ Amazon: 3 orders, eBay: 2 orders │ Action required       │  │
│  │ [Cancel orders] [Find alternative] [Contact customers]   │  │
│  ├──────────────────────────────────────────────────────────┤  │
│  │ SKU-123 "Headphones" - Stock discrepancy detected        │  │
│  │ System: 45 │ Physical count: 42 │ Diff: -3               │  │
│  │ [Adjust inventory] [Investigate] [Ignore]                │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  🟡 Warning (8)                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ SKU-234 - Low stock: 12 units (reorder point: 15)        │  │
│  │ SKU-567 - Slow moving: 0 sales in 30 days                │  │
│  │ SKU-890 - High demand: selling 3x faster than usual      │  │
│  │ ... and 5 more                                            │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

**Alert Configuration:**
```yaml
stock_alerts:
  critical:
    - type: out_of_stock
      action: [email, slack, sms]
      auto_disable_listings: true

    - type: oversold
      action: [email, slack, pagerduty]
      auto_action: hold_orders

    - type: discrepancy
      threshold: 5%
      action: [email]

  warning:
    - type: low_stock
      threshold: reorder_point
      action: [email, slack]
      include_supplier_info: true

    - type: slow_moving
      days_without_sale: 30
      action: [email]
      suggest: markdown_price

    - type: high_demand
      multiplier: 3x
      action: [slack]
      suggest: increase_reorder
```

---

### 5. Multi-Store Connector

Интеграции с популярными платформами.

```
┌─────────────────────────────────────────────────────────────────┐
│  Connected Channels                                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐               │
│  │  Shopify    │ │   Amazon    │ │ WooCommerce │               │
│  │  ─────────  │ │  ─────────  │ │  ─────────  │               │
│  │  ✓ Active   │ │  ✓ Active   │ │  ✓ Active   │               │
│  │  2 stores   │ │  US + EU    │ │  1 store    │               │
│  │  1.2k SKUs  │ │  800 SKUs   │ │  1.2k SKUs  │               │
│  └─────────────┘ └─────────────┘ └─────────────┘               │
│                                                                  │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐               │
│  │    eBay     │ │   Walmart   │ │    Etsy     │               │
│  │  ─────────  │ │  ─────────  │ │  ─────────  │               │
│  │  ✓ Active   │ │  ◐ Setup    │ │  ○ Disabled │               │
│  │  US only    │ │  Pending    │ │  ──         │               │
│  │  450 SKUs   │ │  ──         │ │  ──         │               │
│  └─────────────┘ └─────────────┘ └─────────────┘               │
│                                                                  │
│  ERP / Backend Systems:                                          │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐               │
│  │  1C:Enter   │ │  SAP B1     │ │  QuickBooks │               │
│  │  ─────────  │ │  ─────────  │ │  ─────────  │               │
│  │  ✓ Active   │ │  ○ N/A      │ │  ✓ Active   │               │
│  │  Bi-direct  │ │             │ │  Orders     │               │
│  └─────────────┘ └─────────────┘ └─────────────┘               │
│                                                                  │
│  [+ Add Channel]                                                 │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

**Supported Integrations:**

| Category | Platform | Sync Type | Features |
|----------|----------|-----------|----------|
| **Marketplaces** | Amazon | Bi-directional | Inventory, Orders, FBA |
| | eBay | Bi-directional | Inventory, Orders |
| | Walmart | Bi-directional | Inventory, Orders |
| | Etsy | Bi-directional | Inventory, Orders |
| **E-commerce** | Shopify | Bi-directional | Full sync |
| | WooCommerce | Bi-directional | Full sync |
| | Magento | Bi-directional | Full sync |
| | BigCommerce | Bi-directional | Full sync |
| **ERP** | 1C:Enterprise | Bi-directional | Products, Orders, Stock |
| | SAP Business One | Bi-directional | Products, Orders |
| | NetSuite | Bi-directional | Full ERP sync |
| **Accounting** | QuickBooks | Orders → QB | Orders, Invoices |
| | Xero | Orders → Xero | Orders, Invoices |
| **Shipping** | ShipStation | Orders | Fulfillment |
| | Shippo | Orders | Labels, Tracking |

---

## Technical Architecture

### Project Structure

```
streamline/
├── cmd/
│   └── streamline/
│       └── main.go
├── internal/
│   ├── inventory/
│   │   ├── engine.go           # Inventory sync engine
│   │   ├── allocator.go        # Stock allocation
│   │   ├── reconciler.go       # Discrepancy resolution
│   │   └── forecaster.go       # Demand forecasting
│   ├── orders/
│   │   ├── router.go           # Order routing
│   │   ├── processor.go        # Order processing
│   │   └── fulfillment.go      # Fulfillment tracking
│   ├── pricing/
│   │   ├── engine.go           # Price calculation
│   │   ├── rules.go            # Pricing rules
│   │   └── competitor.go       # Competitive pricing
│   ├── alerts/
│   │   ├── monitor.go          # Stock monitoring
│   │   └── notifier.go         # Alert notifications
│   ├── connectors/
│   │   ├── shopify/
│   │   ├── amazon/
│   │   ├── ebay/
│   │   ├── woocommerce/
│   │   ├── erp/
│   │   │   ├── oneс.go
│   │   │   └── sap.go
│   │   └── connector.go        # Interface
│   └── api/
│       └── handlers.go
└── pkg/
    └── models/
        ├── product.go
        ├── order.go
        └── inventory.go
```

### Data Models

```go
// Core models
type Product struct {
    ID          string            `json:"id"`
    SKU         string            `json:"sku"`
    Name        string            `json:"name"`
    Description string            `json:"description"`
    BasePrice   decimal.Decimal   `json:"base_price"`
    Cost        decimal.Decimal   `json:"cost"`
    Weight      float64           `json:"weight"`
    Dimensions  Dimensions        `json:"dimensions"`
    Variants    []Variant         `json:"variants"`
    Channels    []ChannelListing  `json:"channels"`
    Tags        []string          `json:"tags"`
}

type Inventory struct {
    ProductID    string                     `json:"product_id"`
    TotalStock   int                        `json:"total_stock"`
    Allocated    map[string]int             `json:"allocated"`    // channel -> qty
    Reserved     int                        `json:"reserved"`
    Available    int                        `json:"available"`
    Warehouses   map[string]WarehouseStock  `json:"warehouses"`
    LastUpdated  time.Time                  `json:"last_updated"`
}

type Order struct {
    ID           string        `json:"id"`
    Channel      string        `json:"channel"`
    ChannelRef   string        `json:"channel_ref"`
    Status       OrderStatus   `json:"status"`
    Items        []OrderItem   `json:"items"`
    Customer     Customer      `json:"customer"`
    Shipping     ShippingInfo  `json:"shipping"`
    Totals       OrderTotals   `json:"totals"`
    Warehouse    string        `json:"warehouse"`
    Fulfillment  *Fulfillment  `json:"fulfillment"`
    CreatedAt    time.Time     `json:"created_at"`
}
```

---

## API Endpoints

```yaml
/api/v1/streamline:
  # Products
  GET    /products                    # List products
  POST   /products                    # Create product
  GET    /products/{id}               # Get product
  PUT    /products/{id}               # Update product
  DELETE /products/{id}               # Delete product

  # Inventory
  GET    /inventory                   # Get all inventory
  GET    /inventory/{sku}             # Get SKU inventory
  PUT    /inventory/{sku}             # Update inventory
  POST   /inventory/{sku}/adjust      # Adjust stock
  POST   /inventory/{sku}/transfer    # Transfer between warehouses
  GET    /inventory/low-stock         # Get low stock items

  # Orders
  GET    /orders                      # List orders
  GET    /orders/{id}                 # Get order
  POST   /orders/{id}/route           # Route order to warehouse
  POST   /orders/{id}/fulfill         # Mark as fulfilled
  POST   /orders/{id}/cancel          # Cancel order

  # Channels
  GET    /channels                    # List connected channels
  POST   /channels                    # Connect new channel
  GET    /channels/{id}/sync          # Get sync status
  POST   /channels/{id}/sync          # Force sync

  # Pricing
  GET    /pricing/{sku}               # Get pricing info
  PUT    /pricing/{sku}               # Update pricing
  GET    /pricing/rules               # Get pricing rules
  POST   /pricing/rules               # Create pricing rule

  # Alerts
  GET    /alerts                      # Get active alerts
  POST   /alerts/{id}/acknowledge     # Acknowledge alert
  GET    /alerts/settings             # Get alert settings
  PUT    /alerts/settings             # Update settings
```

---

## Configuration

```yaml
# streamline.yaml
streamline:
  enabled: true

  inventory:
    sync_interval: 30s
    allocation_strategy: weighted
    safety_buffer: 10%
    low_stock_threshold: 15

  orders:
    routing_optimization: cost  # cost, speed, balance
    auto_route: true
    hold_oversold: true

  pricing:
    auto_sync: true
    min_margin: 20%
    competitive_pricing: true

  channels:
    shopify:
      api_key: ${SHOPIFY_API_KEY}
      shop_domain: mystore.myshopify.com
      sync_products: true
      sync_orders: true
      sync_inventory: true

    amazon:
      seller_id: ${AMAZON_SELLER_ID}
      mws_auth_token: ${AMAZON_MWS_TOKEN}
      marketplace_ids: [ATVPDKIKX0DER]  # US
      use_fba: true

    woocommerce:
      url: https://mystore.com
      consumer_key: ${WOO_KEY}
      consumer_secret: ${WOO_SECRET}

  alerts:
    channels:
      slack:
        webhook: ${SLACK_WEBHOOK}
        channel: "#inventory-alerts"
      email:
        recipients: [ops@company.com]
```

---

## Pricing & Packaging

| Feature | Community | Pro | Enterprise |
|---------|-----------|-----|------------|
| Products | 100 | 10,000 | Unlimited |
| Channels | 2 | 5 | Unlimited |
| Warehouses | 1 | 5 | Unlimited |
| Order Routing | Manual | Auto | Auto + Custom |
| Pricing Rules | 3 | 20 | Unlimited |
| Alerts | Basic | Full | Full + Custom |
| ERP Integration | ❌ | Basic | Full |
| API Access | ❌ | ✅ | ✅ |
| Dedicated Support | ❌ | ❌ | ✅ |

---

## Development Status

**Status: Production Ready ✅**

### Core Features (Complete)
- [x] Inventory sync engine
- [x] Shopify connector
- [x] Amazon connector
- [x] WooCommerce connector
- [x] eBay connector
- [x] Order routing
- [x] Price sync
- [x] Real-time alerts
- [x] Dashboard UI

### Future Enhancements
- [ ] 1C / SAP integration
- [ ] Multi-warehouse support
- [ ] Demand forecasting
- [ ] Custom workflows
- [ ] Walmart, Etsy connectors
- [ ] AI-powered pricing
- [ ] Dropship management
- [ ] B2B portal

---

*Document version: 1.0*
*Last updated: November 2024*
