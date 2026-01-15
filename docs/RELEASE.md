# Panduan Release Tag - Asentric SDK

## Cara Membuat Release Tag v0.2.1

### Prerequisites

1. Pastikan semua perubahan sudah di-commit
2. Pastikan semua perubahan sudah di-push ke remote
3. Pastikan versi sudah di-update di semua file yang diperlukan

---

## Step-by-Step Release Process

### 1. Verifikasi Perubahan

```bash
# Cek status git
git status

# Cek perubahan yang akan di-commit
git diff

# Pastikan semua file penting sudah di-commit
git log --oneline -10
```

### 2. Commit Perubahan (jika ada)

```bash
# Jika ada perubahan yang belum di-commit
git add .
git commit -m "chore: prepare release v0.2.1"
```

### 3. Push ke Remote

```bash
# Push semua commit ke remote
git push origin main
# atau
git push origin master
```

### 4. Buat Tag

#### Option A: Lightweight Tag (Simple)

```bash
# Buat lightweight tag
git tag v0.2.1

# Push tag ke remote
git push origin v0.2.1
```

#### Option B: Annotated Tag (Recommended)

```bash
# Buat annotated tag dengan message
git tag -a v0.2.1 -m "Release v0.2.1

Features:
- Default ERC20 rule untuk auto-detect semua Transfer/Mint/Burn events
- Standard ERC20 ABI terintegrasi
- Auto-decode events menggunakan ABI registry
- Console logging untuk demo

See CHANGELOG.md for full details."

# Push tag ke remote
git push origin v0.2.1
```

#### Option C: Tag dengan Signature (Optional)

```bash
# Buat signed tag (jika GPG key sudah setup)
git tag -s v0.2.1 -m "Release v0.2.1"

# Push tag ke remote
git push origin v0.2.1
```

### 5. Push Semua Tags (jika perlu)

```bash
# Push semua tags yang belum di-push
git push --tags
```

---

## Verifikasi Tag

### Cek Tag Lokal

```bash
# List semua tags
git tag

# Cek detail tag
git show v0.2.1

# Cek tag dengan pattern
git tag -l "v0.2*"
```

### Cek Tag di Remote

```bash
# Fetch tags dari remote
git fetch --tags

# List remote tags
git ls-remote --tags origin
```

---

## Membuat GitHub Release (Jika Menggunakan GitHub)

### Via GitHub Web UI

1. Buka repository di GitHub
2. Klik **"Releases"** di sidebar kanan
3. Klik **"Draft a new release"**
4. Pilih tag `v0.2.1` (atau buat tag baru)
5. Isi informasi release:
   - **Title**: `v0.2.1 - Default ERC20 Rule & Auto Decode`
   - **Description**: Copy dari CHANGELOG.md section v0.2.1
6. Upload binaries (jika ada)
7. Klik **"Publish release"**

### Via GitHub CLI (gh)

```bash
# Install GitHub CLI jika belum
# brew install gh (macOS)
# atau download dari https://cli.github.com

# Login ke GitHub
gh auth login

# Buat release dari tag yang sudah ada
gh release create v0.2.1 \
  --title "v0.2.1 - Default ERC20 Rule & Auto Decode" \
  --notes-file CHANGELOG.md \
  --target main
```

### Via GitHub CLI dengan Notes File

```bash
# Buat file release notes
cat > release-notes-v0.2.1.md << 'EOF'
# Release v0.2.1

## ✨ New Features

- **Default ERC20 Rule** - Rule default yang secara otomatis mendeteksi semua event ERC20 Transfer/Mint/Burn dengan nilai > 100
- **Standard ERC20 ABI** - ABI standard ERC20 terintegrasi untuk decode events
- **Auto Decode** - Context builder secara otomatis decode events menggunakan ABI registry
- **Console Logging** - Logging otomatis ke console untuk demo di default rule

## 🔄 Changes

- Context builder sekarang auto-decode events jika ABI registry tersedia
- Template project include default ERC20 rule secara otomatis
- Runtime builder sudah include ABI registry dan context builder

## 🐛 Fixes

- ABI decoding untuk events tanpa registry yang terdaftar
- Console logging format di default rule

## 📖 Documentation

- Added comprehensive guide: `docs/GUIDE-v0.2.1.md`
- Updated template README with new features
- Updated CHANGELOG.md

## 🚀 Quick Start

```bash
asentric init my-watcher
cd my-watcher
go mod tidy
go run cmd/watcher/main.go
```

See [GUIDE-v0.2.1.md](docs/GUIDE-v0.2.1.md) for full documentation.
EOF

# Buat release dengan notes file
gh release create v0.2.1 \
  --title "v0.2.1 - Default ERC20 Rule & Auto Decode" \
  --notes-file release-notes-v0.2.1.md \
  --target main
```

---

## Update Go Modules (Jika Diperlukan)

Jika SDK ini digunakan sebagai dependency di project lain:

```bash
# Di project yang menggunakan SDK
go get github.com/asentric/asentric@v0.2.1
go mod tidy
```

---

## Checklist Release

Sebelum membuat release, pastikan:

- [ ] Versi sudah di-update di:
  - [ ] `cmd/asentric/cmd/version.go`
  - [ ] `cmd/runtime-reference/main.go`
  - [ ] `CHANGELOG.md`
- [ ] Semua perubahan sudah di-commit
- [ ] Semua perubahan sudah di-push ke remote
- [ ] Tests sudah dijalankan dan pass
- [ ] Documentation sudah di-update
- [ ] CHANGELOG.md sudah di-update
- [ ] Tag sudah dibuat dan di-push
- [ ] GitHub release sudah dibuat (jika menggunakan GitHub)

---

## Rollback Tag (Jika Perlu)

### Hapus Tag Lokal

```bash
# Hapus tag lokal
git tag -d v0.2.1
```

### Hapus Tag di Remote

```bash
# Hapus tag di remote
git push origin --delete v0.2.1
# atau
git push origin :refs/tags/v0.2.1
```

---

## Best Practices

### 1. Semantic Versioning

Gunakan [Semantic Versioning](https://semver.org/):
- **MAJOR** (1.0.0): Breaking changes
- **MINOR** (0.2.0): New features, backward compatible
- **PATCH** (0.2.1): Bug fixes, backward compatible

### 2. Tag Naming

- Gunakan format: `v<MAJOR>.<MINOR>.<PATCH>`
- Contoh: `v0.2.1`, `v1.0.0`, `v2.3.4`
- Jangan gunakan prefix lain seperti `release-` atau `version-`

### 3. Release Notes

- Selalu buat release notes yang jelas
- Include breaking changes (jika ada)
- Include migration guide (jika ada)
- Link ke documentation yang relevan

### 4. Testing

- Selalu test release sebelum publish
- Test dengan fresh install
- Test dengan upgrade dari versi sebelumnya

---

## Contoh Release Script

Buat file `scripts/release.sh`:

```bash
#!/bin/bash

set -e

VERSION=$1

if [ -z "$VERSION" ]; then
    echo "Usage: ./scripts/release.sh <version>"
    echo "Example: ./scripts/release.sh 0.2.1"
    exit 1
fi

TAG="v${VERSION}"

echo "🚀 Creating release ${TAG}..."

# Check if tag already exists
if git rev-parse "$TAG" >/dev/null 2>&1; then
    echo "❌ Tag ${TAG} already exists!"
    exit 1
fi

# Check if working directory is clean
if ! git diff-index --quiet HEAD --; then
    echo "❌ Working directory is not clean!"
    echo "Please commit or stash your changes first."
    exit 1
fi

# Create annotated tag
git tag -a "$TAG" -m "Release ${TAG}

See CHANGELOG.md for full details."

# Push tag
echo "📤 Pushing tag to remote..."
git push origin "$TAG"

echo "✅ Release ${TAG} created successfully!"
echo ""
echo "Next steps:"
echo "1. Create GitHub release: gh release create ${TAG}"
echo "2. Or create release via GitHub web UI"
```

Gunakan script:

```bash
chmod +x scripts/release.sh
./scripts/release.sh 0.2.1
```

---

## Troubleshooting

### Tag Tidak Muncul di Remote

```bash
# Fetch tags dari remote
git fetch --tags

# Cek apakah tag ada di remote
git ls-remote --tags origin | grep v0.2.1
```

### Tag Sudah Ada di Remote

Jika tag sudah ada di remote dan ingin di-update:

```bash
# Hapus tag lokal
git tag -d v0.2.1

# Hapus tag di remote
git push origin --delete v0.2.1

# Buat tag baru
git tag -a v0.2.1 -m "Release v0.2.1"

# Push tag baru
git push origin v0.2.1
```

### Conflict dengan Tag yang Sudah Ada

Jika tag sudah ada dan berbeda:

```bash
# Cek commit yang di-tag
git show v0.2.1

# Jika perlu, gunakan force (HATI-HATI!)
git tag -f v0.2.1
git push origin v0.2.1 --force
```

---

## Quick Reference

```bash
# Buat tag
git tag -a v0.2.1 -m "Release v0.2.1"

# Push tag
git push origin v0.2.1

# List tags
git tag

# Show tag
git show v0.2.1

# Delete tag (local)
git tag -d v0.2.1

# Delete tag (remote)
git push origin --delete v0.2.1

# Checkout tag
git checkout v0.2.1
```

---

**Selamat Release! 🎉**
