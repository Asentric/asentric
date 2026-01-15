# Quick Demo Guide - 5 Menit

## 🎯 Tujuan

Menunjukkan SDK secara otomatis mendeteksi semua ERC20 events tanpa konfigurasi tambahan.

---

## ⚡ Quick Start (30 detik)

```bash
# 1. Install CLI
go install github.com/asentric/asentric/cmd/asentric@latest

# 2. Buat project
asentric init demo
cd demo
go mod tidy

# 3. Jalankan
go run cmd/watcher/main.go
```

---

## 📺 Script Presentasi

### Opening
> "Asentric SDK v0.2.1 - Real-time blockchain monitoring dengan zero configuration. Mari kita lihat..."

### Setup (30 detik)
> "3 command saja: init, cd, go mod tidy. Project sudah siap dengan default rule."

### Running (30 detik)
> "Jalankan watcher. Perhatikan rule default sudah ter-register. Siap mendengarkan semua ERC20 events."

### Live Demo (2-3 menit)
> "Tunggu events... [Saat event muncul] Lihat! Event terdeteksi otomatis:
> - Contract address
> - From/To addresses  
> - Value
> - Block & Tx hash
> 
> Semua ini tanpa konfigurasi tambahan!"

### Highlight (1 menit)
> "Yang menarik:
> ✅ Zero config - semua ERC20 auto-detect
> ✅ Auto decode - standard ABI terintegrasi
> ✅ Smart filter - hanya > 100 tokens
> ✅ Real-time - langsung muncul di console"

### Closing
> "SDK siap untuk production. Template sudah include default rule. Terima kasih!"

---

## 🎬 Expected Output

```
===========================================
  demo - Asentric Watcher
===========================================
✓ Rules registered
  - erc20-default: Detects all ERC20 Transfer/Mint/Burn events (value > 100)
Connecting to Base Sepolia...
✓ Runtime ready
-------------------------------------------
Chain:  Base Sepolia (ID: 84532)
Source: websocket
Sink:   console
-------------------------------------------
Listening for events... (Press Ctrl+C to stop)

[ERC20 Default] Transfer detected:
  Contract: 0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48
  From: 0x1234...5678
  To: 0xabcd...ef01
  Value: 1000.5
  Block: 12345678
  TxHash: 0xabc123...

[ALERT] erc20-default
  Title: ERC20 Transfer: 1000.5
  Severity: medium
```

---

## ✅ Key Points

1. **Zero Configuration** - Tidak perlu setup tambahan
2. **Auto-Detection** - Semua ERC20 terdeteksi
3. **Real-time** - Events muncul langsung
4. **Extensible** - Mudah tambah custom rules

---

## 🔧 Troubleshooting

**Tidak ada events?**
- Cek RPC endpoint aktif
- Gunakan testnet yang lebih aktif
- Trigger test transaction

**Connection error?**
- Cek RPC URL valid
- Try different endpoint

---

**Selamat Demo! 🚀**
