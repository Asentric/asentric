# Quick Start - Release Tag v0.2.1

## 🚀 Cara Cepat Membuat Release

### 1. Commit Semua Perubahan

```bash
# Cek status
git status

# Add semua perubahan
git add .

# Commit
git commit -m "chore: prepare release v0.2.1"
```

### 2. Push ke Remote

```bash
git push origin main
```

### 3. Buat Tag (Pilih salah satu)

#### Option A: Menggunakan Script (Recommended)

```bash
./scripts/release.sh 0.2.1
```

#### Option B: Manual

```bash
# Buat annotated tag
git tag -a v0.2.1 -m "Release v0.2.1

Features:
- Default ERC20 rule untuk auto-detect semua Transfer/Mint/Burn events
- Standard ERC20 ABI terintegrasi
- Auto-decode events menggunakan ABI registry
- Console logging untuk demo

See CHANGELOG.md for full details."

# Push tag
git push origin v0.2.1
```

### 4. Buat GitHub Release (Opsional)

#### Via GitHub CLI

```bash
gh release create v0.2.1 \
  --title "v0.2.1 - Default ERC20 Rule & Auto Decode" \
  --notes-file CHANGELOG.md
```

#### Via Web UI

1. Buka: https://github.com/YOUR_USERNAME/asentric-sdk/releases/new
2. Pilih tag: `v0.2.1`
3. Title: `v0.2.1 - Default ERC20 Rule & Auto Decode`
4. Description: Copy dari CHANGELOG.md section v0.2.1
5. Klik "Publish release"

---

## ✅ Checklist

Sebelum release, pastikan:

- [ ] Versi sudah di-update di `cmd/asentric/cmd/version.go`
- [ ] Versi sudah di-update di `cmd/runtime-reference/main.go`
- [ ] CHANGELOG.md sudah di-update
- [ ] Semua perubahan sudah di-commit
- [ ] Semua perubahan sudah di-push
- [ ] Tag sudah dibuat dan di-push
- [ ] GitHub release sudah dibuat (opsional)

---

## 📖 Dokumentasi Lengkap

Lihat [docs/RELEASE.md](docs/RELEASE.md) untuk panduan lengkap.

---

**Selamat Release! 🎉**
