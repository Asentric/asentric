# Changelog

All notable changes to this project will be documented in this file.

## [0.2.1] - 2026-01-15

### Added
- ✨ **Default ERC20 Rule** - Rule default yang secara otomatis mendeteksi semua event ERC20 Transfer/Mint/Burn dengan nilai > 100
- ✨ **Standard ERC20 ABI** - ABI standard ERC20 terintegrasi untuk decode events
- ✨ **Auto Decode** - Context builder secara otomatis decode events menggunakan ABI registry
- ✨ **Console Logging** - Logging otomatis ke console untuk demo di default rule

### Changed
- 🔄 Context builder sekarang auto-decode events jika ABI registry tersedia
- 🔄 Template project include default ERC20 rule secara otomatis
- 🔄 Runtime builder sudah include ABI registry dan context builder

### Fixed
- 🐛 ABI decoding untuk events tanpa registry yang terdaftar
- 🐛 Console logging format di default rule

---

## [0.2.0] - 2026-01-15

### Changed

- **Default Network**: Changed default configuration from Base Sepolia to Mantle Sepolia (Chain ID: 5003)
- **RPC Endpoints**: Updated default WebSocket RPC to `wss://mantle-sepolia.drpc.org`
- **Go Version**: Updated minimum Go version to 1.22
- **Detection Threshold**: Lowered default threshold to 1 token for demo visibility
- **Documentation**: Complete rewrite in professional English for hackathon submission

### Added

- **Quick Start Guide**: New `docs/QUICK-START.md` for rapid onboarding
- **ERC20 ABI**: Added default ERC20 ABI file in templates
- **Demo Guide**: Comprehensive demo documentation at `test/DEMO.md`

### Fixed

- WebSocket RPC endpoint reliability (switched to dRPC)
- Template README formatting and structure

### Removed

- Redis requirement for basic usage
- Emojis from documentation

---

## [0.1.1] - 2025-12-XX

### Added

- Initial CLI with `init` command
- Project scaffolding templates
- Example detection rules

---

## [0.1.0] - 2025-12-XX

### Added

- Core SDK with Engine, Rule, and Context
- WebSocket event source
- Console and Webhook alert sinks
- Basic documentation
