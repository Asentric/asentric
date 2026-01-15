# Release v0.2.2 - Quick Guide

## 🚀 Release Checklist

### 1. Commit Semua Perubahan

```bash
cd /Users/macbookpro/Documents/Projects/Asentric/asentric-sdk

# Add semua perubahan
git add .

# Commit
git commit -m "chore: prepare release v0.2.2

Fixes:
- CLI install path fix (cmd/asentric)
- Type mismatch fix in context_builder.go
- Template version update to v0.2.2
- Documentation updates with correct install path

See CHANGELOG.md for details."
```

### 2. Push ke Remote

```bash
git push origin main
```

### 3. Buat Tag

```bash
# Buat annotated tag
git tag -a v0.2.2 -m "Release v0.2.2

Fixes:
- CLI install path fix (cmd/asentric)
- Type mismatch fix in context_builder.go  
- Template version update to v0.2.2
- Documentation updates

See CHANGELOG.md for full details."

# Push tag
git push origin v0.2.2
```

### 4. Verifikasi Tag

```bash
# Cek tag
git show v0.2.2

# List tags
git tag -l "v0.2*"
```

### 5. Buat GitHub Release (Opsional)

```bash
# Via GitHub CLI
gh release create v0.2.2 \
  --title "v0.2.2 - Bug Fixes & Install Path Correction" \
  --notes-file CHANGELOG.md \
  --target main
```

Atau via web UI: https://github.com/YOUR_USERNAME/asentric-sdk/releases/new

---

## ✅ Perubahan di v0.2.2

### Fixed
- 🐛 **CLI Install Path** - Fixed install command
- 🐛 **Type Mismatch** - Fixed context_builder.go
- 🐛 **Template Version** - Updated to v0.2.2

### Changed
- 🔄 Updated documentation with correct install path
- 🔄 Template uses v0.2.2 by default

---

## 📝 Quick Commands

```bash
# 1. Commit & Push
git add .
git commit -m "chore: prepare release v0.2.2"
git push origin main

# 2. Create Tag
git tag -a v0.2.2 -m "Release v0.2.2"
git push origin v0.2.2

# 3. Create Release (optional)
gh release create v0.2.2 --title "v0.2.2" --notes-file CHANGELOG.md
```

---

**Selamat Release! 🎉**
