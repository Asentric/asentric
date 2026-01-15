# Quick Start Guide

Get started with Asentric SDK in under 5 minutes.

---

## Prerequisites

- Go 1.22 or higher

---

## Step 1: Install CLI

```bash
go install github.com/asentric/asentric@latest
```

Verify installation:

```bash
asentric --version
```

---

## Step 2: Create Project

```bash
asentric init my-watcher
cd my-watcher
go mod tidy
```

---

## Step 3: Run

```bash
go run cmd/watcher/main.go
```

The watcher connects to Mantle Sepolia by default and starts monitoring.

---

## Step 4: Configure (Optional)

### Change Network

Edit `config/asentric.yaml`:

```yaml
chain:
  id: 84532                              # Base Sepolia
  name: "Base Sepolia"
  rpcWs: "wss://base-sepolia.drpc.org"

source:
  url: "wss://base-sepolia.drpc.org"
```

### Add Target Contract

Edit `config/registry.yaml`:

```yaml
targets:
  - address: "0xYourContractAddress"
    name: "My Token"
    abi_path: "abi/erc20.json"
```

### Enable Webhook

Edit `config/asentric.yaml`:

```yaml
sink:
  type: "webhook"
  url: "http://localhost:8080/api/webhook"
```

---

## Step 5: Write Custom Rule

Create `rules/my_rule.go`:

```go
package rules

import (
    "github.com/asentric/asentric/pkg/asentric"
    "github.com/asentric/asentric/pkg/utils"
)

type MyRule struct{}

func (r *MyRule) Name() string {
    return "my-rule"
}

func (r *MyRule) Severity() asentric.Severity {
    return asentric.SeverityMedium
}

func (r *MyRule) Evaluate(ctx asentric.Context) (*asentric.Alert, error) {
    for _, log := range ctx.Logs() {
        if log.Event.Name == "Transfer" {
            from := utils.GetFieldString(log.Event.Fields, "from")
            
            if utils.IsZeroAddress(from) {
                return asentric.NewAlert(
                    r.Name(),
                    "Mint Detected",
                    r.Severity(),
                ), nil
            }
        }
    }
    return nil, nil
}
```

Register in `cmd/watcher/main.go`:

```go
engine.RegisterRule(&rules.MyRule{})
```

---

## Supported Networks

| Network | Chain ID | WebSocket RPC |
|---------|----------|---------------|
| Mantle Sepolia | 5003 | `wss://mantle-sepolia.drpc.org` |
| Base Sepolia | 84532 | `wss://base-sepolia.drpc.org` |
| Ethereum Sepolia | 11155111 | `wss://eth-sepolia.drpc.org` |

---

## Next Steps

- [Developer Overview](developer-overview.md) - Complete guide
- [SDK API Reference](sdk-api.md) - API documentation
- [Architecture](architecture.md) - System design
