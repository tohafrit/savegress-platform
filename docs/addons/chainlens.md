# ChainLens — Blockchain & Web3 Add-on

> "See through your smart contracts"

## Overview

ChainLens — инструмент для разработчиков Web3, который объединяет анализ смарт-контрактов, отладку транзакций и мониторинг on-chain событий. В связке с Savegress CDC обеспечивает полную картину Web3 бизнеса.

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           CHAINLENS                                      │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │                    VS Code Extension                             │   │
│   │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐        │   │
│   │  │ Contract │  │Transaction│  │   Gas    │  │ Security │        │   │
│   │  │ Analyzer │  │  Tracer  │  │ Profiler │  │ Scanner  │        │   │
│   │  └──────────┘  └──────────┘  └──────────┘  └──────────┘        │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │                    Web Dashboard                                 │   │
│   │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐        │   │
│   │  │ Contract │  │  Event   │  │  Wallet  │  │  Alert   │        │   │
│   │  │ Monitor  │  │  Stream  │  │ Tracker  │  │  System  │        │   │
│   │  └──────────┘  └──────────┘  └──────────┘  └──────────┘        │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │              Savegress CDC Integration                           │   │
│   │                                                                  │   │
│   │   PostgreSQL ◄──CDC──► ChainLens ◄──RPC──► Ethereum/Polygon     │   │
│   │   (off-chain)                              (on-chain)            │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Target Audience

| Segment | Use Case | Pain Point Solved |
|---------|----------|-------------------|
| **DeFi Protocols** | Contract monitoring | Security + debugging |
| **NFT Marketplaces** | Royalty tracking | On-chain event sync |
| **Crypto Exchanges** | Wallet monitoring | Transaction auditing |
| **GameFi Studios** | In-game asset tracking | Cross-chain sync |
| **DAOs** | Governance tracking | Vote/proposal monitoring |

---

## Core Features

### 1. Smart Contract Analyzer

Статический анализ Solidity кода в VS Code.

```
┌─────────────────────────────────────────────────────────────────┐
│  Contract: NFTMarketplace.sol                      ChainLens    │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  45│  function buyNFT(uint256 tokenId) external payable {       │
│  46│      Listing memory listing = listings[tokenId];           │
│  47│      require(listing.active, "Not listed");                │
│  48│      require(msg.value >= listing.price, "Low price");     │
│    │      ⚠️ Reentrancy risk: state change after transfer       │
│  49│                                                             │
│  50│      payable(listing.seller).transfer(msg.value);          │
│    │      💰 Gas: ~21,000 (consider using call)                 │
│  51│      listings[tokenId].active = false;                     │
│    │      ⚠️ State change after external call                   │
│  52│      nft.transferFrom(address(this), msg.sender, tokenId); │
│  53│  }                                                          │
│                                                                  │
│  ─────────────────────────────────────────────────────────────  │
│  Issues: 2 warnings │ Gas estimate: 45,230 │ Complexity: Medium │
└─────────────────────────────────────────────────────────────────┘
```

**Security Checks:**
```yaml
security_rules:
  critical:
    - reentrancy
    - integer_overflow
    - unchecked_return
    - delegatecall_injection
    - selfdestruct_exposure

  warning:
    - floating_pragma
    - missing_zero_check
    - state_after_external
    - unbounded_loop
    - timestamp_dependence

  info:
    - gas_optimization
    - naming_convention
    - missing_natspec
```

---

### 2. Transaction Tracer

Пошаговая отладка транзакций с привязкой к исходному коду.

```
┌─────────────────────────────────────────────────────────────────┐
│  Transaction Trace: 0x7a3f...8b2c                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─ CALL NFTMarketplace.buyNFT(tokenId: 1234)                   │
│  │  From: 0xBuyer...                                            │
│  │  Value: 1.5 ETH                                              │
│  │  Gas: 150,000                                                │
│  │                                                               │
│  │  ├─ SLOAD listings[1234]                    // 2,100 gas     │
│  │  │  → {seller: 0xSeller, price: 1.5 ETH, active: true}      │
│  │  │                                                           │
│  │  ├─ CALL 0xSeller.transfer(1.5 ETH)         // 21,000 gas   │
│  │  │  └─ SUCCESS                                               │
│  │  │                                                           │
│  │  ├─ SSTORE listings[1234].active = false    // 5,000 gas    │
│  │  │                                                           │
│  │  └─ CALL NFT.transferFrom(...)              // 45,000 gas   │
│  │      ├─ SLOAD ownerOf[1234]                                  │
│  │      ├─ SSTORE ownerOf[1234] = 0xBuyer                       │
│  │      └─ LOG Transfer(from, to, tokenId)                      │
│  │                                                               │
│  └─ SUCCESS                                                      │
│                                                                  │
│  Total Gas: 73,100 │ Status: Success │ Block: 18,542,301        │
└─────────────────────────────────────────────────────────────────┘
```

**Trace Features:**
- Step-by-step execution
- Storage reads/writes
- Internal calls visualization
- Gas breakdown per operation
- Revert reason decoding
- Event log highlighting

---

### 3. Gas Profiler

Оптимизация газа с рекомендациями.

```
┌─────────────────────────────────────────────────────────────────┐
│  Gas Profile: NFTMarketplace                                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Function              Avg Gas    Calls    Total Cost (30 gwei) │
│  ─────────────────────────────────────────────────────────────  │
│  buyNFT()              73,100     1,234    $2,701.42            │
│  listNFT()             45,200       892    $1,210.78            │
│  cancelListing()       22,100       156      $103.45            │
│  updatePrice()         28,400       445      $379.34            │
│                                                                  │
│  ─────────────────────────────────────────────────────────────  │
│  Total (30 days)                           $4,394.99            │
│                                                                  │
│  Optimization Suggestions:                                       │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ 1. buyNFT(): Use `call` instead of `transfer`            │  │
│  │    Potential savings: ~2,300 gas per call                │  │
│  │    Monthly savings: ~$85                                  │  │
│  │                                                           │  │
│  │ 2. listNFT(): Pack struct variables                      │  │
│  │    Potential savings: ~5,000 gas per call                │  │
│  │    Monthly savings: ~$134                                 │  │
│  │                                                           │  │
│  │ 3. Use immutable for constant addresses                  │  │
│  │    Potential savings: ~200 gas per read                  │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

### 4. Contract Monitor

Real-time мониторинг deployed контрактов.

```
┌─────────────────────────────────────────────────────────────────┐
│  Contract Monitor                                    [Live]     │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  NFTMarketplace (0x1234...5678)                                 │
│  ├─ Status: Active ✓                                            │
│  ├─ Balance: 12.5 ETH                                           │
│  ├─ Events (24h): 1,247                                         │
│  └─ Alerts: 0                                                    │
│                                                                  │
│  Recent Events:                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ 14:32:15  NFTListed    tokenId=5678 price=2.5 ETH       │  │
│  │ 14:31:02  NFTSold      tokenId=1234 buyer=0xAbc...      │  │
│  │ 14:28:45  PriceUpdated tokenId=9012 newPrice=1.8 ETH    │  │
│  │ 14:25:11  NFTSold      tokenId=3456 buyer=0xDef...      │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  Alert Rules:                                                    │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ ✓ Large transfer (> 10 ETH)         → Slack + Email      │  │
│  │ ✓ Failed transaction                → PagerDuty          │  │
│  │ ✓ Unusual gas spike                 → Slack              │  │
│  │ ✓ Contract balance low (< 1 ETH)    → Email              │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

### 5. Savegress CDC Integration

Синхронизация on-chain и off-chain данных.

```
┌─────────────────────────────────────────────────────────────────┐
│                    CDC + ChainLens Integration                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   Off-chain (PostgreSQL)              On-chain (Ethereum)       │
│   ┌─────────────────┐                ┌─────────────────┐        │
│   │ users           │                │ NFTMarketplace  │        │
│   │ ├─ id           │                │ ├─ listings[]   │        │
│   │ ├─ wallet       │◄─── match ────►│ ├─ owners[]     │        │
│   │ └─ email        │                │ └─ events       │        │
│   │                 │                │                 │        │
│   │ orders          │                │ PaymentSplitter │        │
│   │ ├─ id           │                │ ├─ shares[]     │        │
│   │ ├─ nft_id       │◄─── sync ─────►│ ├─ released[]   │        │
│   │ └─ price        │                │ └─ events       │        │
│   └─────────────────┘                └─────────────────┘        │
│           │                                   │                  │
│           └──────────────┬───────────────────┘                  │
│                          │                                       │
│                          ▼                                       │
│              ┌─────────────────────┐                            │
│              │   Unified View      │                            │
│              │                     │                            │
│              │  User: alice@...    │                            │
│              │  Wallet: 0xAbc...   │                            │
│              │  NFTs owned: 12     │                            │
│              │  Total value: 45 ETH│                            │
│              │  Royalties: 2.3 ETH │                            │
│              └─────────────────────┘                            │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

**Use Cases:**
- NFT ownership sync with user profiles
- Transaction history enrichment
- Royalty tracking and reporting
- Wallet balance monitoring
- Cross-chain activity correlation

---

## Technical Architecture

### Project Structure

```
chainlens/
├── vscode-extension/
│   ├── src/
│   │   ├── extension.ts          # Entry point
│   │   ├── analyzer/
│   │   │   ├── solidity.ts       # Solidity parser
│   │   │   ├── security.ts       # Security rules
│   │   │   └── gas.ts            # Gas estimation
│   │   ├── tracer/
│   │   │   ├── debugger.ts       # Transaction tracer
│   │   │   └── decoder.ts        # ABI decoder
│   │   ├── providers/
│   │   │   ├── ethereum.ts       # Ethereum RPC
│   │   │   └── sourcify.ts       # Contract verification
│   │   └── views/
│   │       ├── trace.ts          # Trace view
│   │       └── analysis.ts       # Analysis view
│   └── package.json
│
├── backend/
│   ├── cmd/chainlens/
│   │   └── main.go
│   ├── internal/
│   │   ├── monitor/
│   │   │   ├── contract.go       # Contract monitoring
│   │   │   ├── events.go         # Event indexing
│   │   │   └── alerts.go         # Alert triggers
│   │   ├── indexer/
│   │   │   ├── ethereum.go       # Ethereum indexer
│   │   │   ├── polygon.go        # Polygon indexer
│   │   │   └── arbitrum.go       # Arbitrum indexer
│   │   ├── sync/
│   │   │   └── cdc_bridge.go     # Savegress CDC bridge
│   │   └── api/
│   │       └── handlers.go
│   └── pkg/
│       └── abi/                  # ABI utilities
│
└── frontend/
    └── dashboard/                # Web dashboard components
```

### Supported Networks

```go
type Network struct {
    Name      string
    ChainID   int64
    RPC       string
    Explorer  string
    Supported bool
}

var SupportedNetworks = []Network{
    {Name: "Ethereum", ChainID: 1, Supported: true},
    {Name: "Polygon", ChainID: 137, Supported: true},
    {Name: "Arbitrum", ChainID: 42161, Supported: true},
    {Name: "Optimism", ChainID: 10, Supported: true},
    {Name: "BSC", ChainID: 56, Supported: true},
    {Name: "Avalanche", ChainID: 43114, Supported: true},
    {Name: "Base", ChainID: 8453, Supported: true},
    // Testnets
    {Name: "Goerli", ChainID: 5, Supported: true},
    {Name: "Sepolia", ChainID: 11155111, Supported: true},
    {Name: "Mumbai", ChainID: 80001, Supported: true},
}
```

---

## API Endpoints

```yaml
/api/v1/chainlens:
  # Contracts
  GET  /contracts                     # List monitored contracts
  POST /contracts                     # Add contract to monitor
  GET  /contracts/{address}           # Get contract details
  DELETE /contracts/{address}         # Remove from monitoring

  # Events
  GET  /contracts/{address}/events    # Get contract events
  GET  /events/stream                 # WebSocket event stream

  # Analysis
  POST /analyze                       # Analyze contract code
  GET  /analyze/{address}             # Get cached analysis

  # Transactions
  GET  /tx/{hash}                     # Get transaction details
  GET  /tx/{hash}/trace               # Get transaction trace

  # Gas
  GET  /contracts/{address}/gas       # Gas profile for contract
  GET  /gas/estimate                  # Estimate gas for call

  # Alerts
  GET  /alerts                        # List alert rules
  POST /alerts                        # Create alert rule
  DELETE /alerts/{id}                 # Delete alert rule

  # Sync (CDC Integration)
  POST /sync/configure                # Configure CDC sync
  GET  /sync/status                   # Get sync status
```

---

## Configuration

```yaml
# chainlens.yaml
chainlens:
  enabled: true

  networks:
    ethereum:
      rpc_url: ${ETHEREUM_RPC_URL}
      ws_url: ${ETHEREUM_WS_URL}
      api_key: ${ETHERSCAN_API_KEY}

    polygon:
      rpc_url: ${POLYGON_RPC_URL}
      ws_url: ${POLYGON_WS_URL}
      api_key: ${POLYGONSCAN_API_KEY}

  analyzer:
    security_rules: all
    gas_estimation: true
    auto_verify: true  # Auto-fetch verified source

  monitor:
    poll_interval: 12s  # Block time
    event_retention: 90d
    max_contracts: 100  # Per user

  alerts:
    channels:
      slack:
        webhook_url: ${SLACK_WEBHOOK}
      discord:
        webhook_url: ${DISCORD_WEBHOOK}

  cdc_integration:
    enabled: true
    sync_events: true
    sync_balances: true
```

---

## Pricing & Packaging

| Feature | Community | Pro | Enterprise |
|---------|-----------|-----|------------|
| VS Code Extension | ✅ | ✅ | ✅ |
| Contract Analysis | 5/day | Unlimited | Unlimited |
| Transaction Traces | 10/day | Unlimited | Unlimited |
| Contract Monitoring | 2 | 20 | Unlimited |
| Networks | 2 | All | All + Private |
| Event Retention | 7d | 90d | Custom |
| CDC Integration | ❌ | ✅ | ✅ |
| Custom RPC | ❌ | ✅ | ✅ |
| Priority Support | ❌ | ❌ | ✅ |

---

## Development Status

### Completed ✅
- [x] VS Code extension core
- [x] Solidity analyzer
- [x] Transaction tracer
- [x] Basic security rules
- [x] Gas profiler

### In Progress 🔨
- [ ] Web dashboard
- [ ] Multi-chain support
- [ ] Alert system
- [ ] CDC integration

### Planned 📋
- [ ] Advanced ML security detection
- [ ] Cross-chain tracking
- [ ] Governance monitoring
- [ ] DeFi protocol templates

---

*Document version: 1.0*
*Last updated: November 2024*
