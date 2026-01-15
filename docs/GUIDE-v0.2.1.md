# Panduan Asentric SDK v0.2.1

## Versi 0.2.1 - Fitur Baru

### ✨ Fitur Utama

1. **Default ERC20 Rule** - Rule default yang secara otomatis mendeteksi semua event ERC20 Transfer/Mint/Burn dengan nilai > 100
2. **Standard ERC20 ABI** - ABI standard ERC20 terintegrasi untuk decode events
3. **Console Logging** - Logging otomatis ke console untuk demo
4. **Auto Decode** - Context builder secara otomatis decode events menggunakan ABI registry

---

## Quick Start

### 1. Install CLI

```bash
go install github.com/asentric/asentric@latest
```

### 2. Buat Project Baru

```bash
asentric init my-watcher
cd my-watcher
go mod tidy
```

### 3. Jalankan Watcher

```bash
go run cmd/watcher/main.go
```

**Watcher akan secara otomatis:**
- ✅ Mendengarkan semua event ERC20 Transfer/Mint/Burn
- ✅ Filter events dengan nilai > 100 tokens
- ✅ Print ke console untuk demo
- ✅ Generate alerts yang bisa dikirim ke webhook

---

## Konfigurasi

### config/asentric.yaml

```yaml
chain:
  id: 84532                    # Base Sepolia
  name: "Base Sepolia"
  rpcWs: "wss://base-sepolia.drpc.org"

source:
  type: "websocket"

sink:
  type: "console"               # atau "webhook"
  # url: "http://localhost:8080/webhook"  # uncomment untuk webhook
```

### config/registry.yaml (Opsional)

Untuk contract spesifik, tambahkan ke registry:

```yaml
targets:
  - address: "0xYourContractAddress"
    name: "My Token"
    abi_path: "abi/erc20.json"
```

**Catatan:** Default rule tidak memerlukan registry untuk ERC20 standard, tapi registry diperlukan untuk contract custom events.

---

## Default Rule - ERC20 Default

### Apa yang Dideteksi?

Rule default (`erc20-default`) akan mendeteksi:

1. **Transfer Events** - Transfer token dari satu address ke address lain
2. **Mint Events** - Mint token baru (from = zero address)
3. **Burn Events** - Burn token (to = zero address)

### Filter

- ✅ Nilai > 100 tokens (default 18 decimals)
- ✅ Semua contract ERC20 (tidak perlu daftar manual)
- ✅ Auto-detect menggunakan standard ERC20 ABI

### Output Console

Ketika event terdeteksi, console akan menampilkan:

```
[ERC20 Default] MINT detected:
  Contract: 0x1234...5678
  From: 0x0000...0000
  To: 0xabcd...ef01
  Value: 1000.5
  Block: 12345
  TxHash: 0xabc123...
```

---

## Custom Rules

### Membuat Rule Baru

Buat file `rules/my_rule.go`:

```go
package rules

import (
    "github.com/asentric/asentric/pkg/asentric"
    "github.com/asentric/asentric/pkg/utils"
)

type MyCustomRule struct{}

func (r *MyCustomRule) Name() string {
    return "my-custom-rule"
}

func (r *MyCustomRule) Severity() asentric.Severity {
    return asentric.SeverityHigh
}

func (r *MyCustomRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    for _, log := range ctx.Logs() {
        if log.Event.Name == "Transfer" {
            value := utils.GetFieldBigInt(log.Event.Fields, "value")
            // Your custom logic here
            if value.Cmp(big.NewInt(1000000000000000000)) > 0 { // 1 token
                return asentric.NewAlert(
                    r.Name(),
                    "Custom Alert",
                    r.Severity(),
                ), nil
            }
        }
    }
    return nil, nil
}
```

### Register di `cmd/watcher/main.go`

```go
engine.RegisterRule(&rules.MyCustomRule{})
```

---

## Built-in Rules

SDK menyediakan beberapa built-in rules:

### 1. ERC20DefaultRule

```go
rules.NewERC20DefaultRule()
```

- Detects: Semua ERC20 Transfer/Mint/Burn
- Threshold: > 100 tokens (default 18 decimals)
- Scope: Semua contract ERC20

### 2. ERC20TransferRule

```go
rules.NewERC20TransferRule(rules.ERC20TransferConfig{
    ContractAddress: "0x...",
    TokenSymbol:     "USDC",
    Decimals:        6,
})
```

- Detects: Transfer events dari contract spesifik
- Includes: Mint dan Burn detection

### 3. LargeTransferRule

```go
rules.NewLargeTransferRule()
```

- Detects: Large native token transfers
- Default threshold: 1000 tokens

### 4. WhaleAlertRule

```go
rules.NewWhaleAlertRule()
```

- Detects: Very large native token transfers
- Default threshold: 10 tokens

---

## ABI Registry

### Standard ERC20 ABI

SDK secara otomatis menggunakan standard ERC20 ABI untuk decode events. Tidak perlu konfigurasi tambahan.

### Custom ABI

Untuk contract custom, tambahkan ABI ke registry:

1. Simpan ABI file di `abi/mycontract.json`
2. Tambahkan ke `config/registry.yaml`:

```yaml
targets:
  - address: "0xMyContract"
    name: "My Contract"
    abi_path: "abi/mycontract.json"
```

---

## Webhook Configuration

### Setup Webhook

1. Edit `config/asentric.yaml`:

```yaml
sink:
  type: "webhook"
  url: "http://localhost:8080/webhook"
```

2. Webhook akan menerima POST request dengan format:

```json
{
  "rule": "erc20-default",
  "severity": "medium",
  "title": "ERC20 Transfer: 1000.5",
  "description": "Transfer from 0x1234...5678 to 0xabcd...ef01",
  "timestamp": "2026-01-15T10:30:00Z",
  "metadata": {
    "from": "0x1234...5678",
    "to": "0xabcd...ef01",
    "value": "1000.5",
    "value_wei": "1000500000000000000000",
    "isMint": false,
    "isBurn": false,
    "contract": "0x..."
  },
  "ref": {
    "txHash": "0xabc123...",
    "blockNumber": 12345,
    "logIndex": 0
  }
}
```

---

## Troubleshooting

### Events Tidak Terdeteksi

1. **Cek Chain ID** - Pastikan chain ID sesuai dengan network
2. **Cek RPC URL** - Pastikan WebSocket RPC endpoint valid
3. **Cek Filter** - Default rule hanya detect events dengan value > 100
4. **Cek ABI** - Pastikan contract menggunakan standard ERC20 interface

### Console Tidak Menampilkan Log

- Pastikan rule sudah ter-register
- Cek apakah ada events dengan value > 100
- Enable debug logging di config:

```yaml
debug: true
```

### Webhook Tidak Menerima Alerts

- Cek URL webhook valid dan accessible
- Test dengan curl:

```bash
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{"test": "data"}'
```

---

## Best Practices

### 1. Production Setup

- ✅ Gunakan webhook untuk production
- ✅ Setup monitoring untuk webhook endpoint
- ✅ Gunakan dedicated RPC endpoint (bukan public)
- ✅ Implement retry logic untuk webhook

### 2. Rule Development

- ✅ Test rules dengan testnet dulu
- ✅ Gunakan console sink untuk development
- ✅ Log semua metadata yang diperlukan
- ✅ Handle edge cases (zero address, large numbers)

### 3. Performance

- ✅ Filter events di rule level (jangan process semua)
- ✅ Gunakan threshold yang reasonable
- ✅ Monitor memory usage untuk long-running processes

---

## Changelog v0.2.1

### Added
- ✨ Default ERC20 rule untuk auto-detect semua Transfer/Mint/Burn events
- ✨ Standard ERC20 ABI terintegrasi
- ✨ Auto-decode events menggunakan ABI registry
- ✨ Console logging untuk demo

### Changed
- 🔄 Context builder sekarang auto-decode events jika ABI tersedia
- 🔄 Template project include default ERC20 rule

### Fixed
- 🐛 ABI decoding untuk events tanpa registry
- 🐛 Console logging format

---

## Support

- 📖 [Full Documentation](https://github.com/asentric/asentric)
- 💬 [Issues](https://github.com/asentric/asentric/issues)
- 📧 Email: support@asentric.io

---

**Happy Monitoring! 🚀**
