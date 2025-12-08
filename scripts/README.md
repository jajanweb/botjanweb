# 🛠️ Scripts Directory

This folder contains utility scripts for development, testing, and maintenance.

---

## 📜 Available Scripts

### 🔒 Security Testing
- **[test_security.sh](test_security.sh)** - Automated security testing suite
  - Tests header authentication
  - Tests rate limiting
  - Tests webhook validation
  - Tests audit logging
  - Usage: `./scripts/test_security.sh`

### 🔄 Migration & Refactoring
- **[migrate_structure.sh](migrate_structure.sh)** - Project structure migration script
  - Moves files to clean architecture structure
  - Updates import paths
  - Usage: `./scripts/migrate_structure.sh`

- **[update_imports.sh](update_imports.sh)** - Import path updater
  - Updates Go import paths after refactoring
  - Usage: `./scripts/update_imports.sh`

---

## 🚀 Usage

### Make Scripts Executable

```bash
chmod +x scripts/*.sh
```

### Run Security Tests

```bash
# Terminal 1: Start bot
make dev

# Terminal 2: Run tests
./scripts/test_security.sh
```

### Run Migration Scripts

```bash
# Migrate project structure
./scripts/migrate_structure.sh

# Update import paths
./scripts/update_imports.sh
```

---

## 📝 Script Guidelines

When adding new scripts:

1. **Naming**: Use lowercase with underscores (e.g., `my_script.sh`)
2. **Shebang**: Always start with `#!/bin/bash`
3. **Documentation**: Add comments explaining what script does
4. **Error Handling**: Use `set -e` to exit on errors
5. **Colors**: Use color codes for better readability
6. **Permissions**: Make executable with `chmod +x`

### Template

```bash
#!/bin/bash
# Script Name: my_script.sh
# Description: What this script does
# Usage: ./scripts/my_script.sh [options]

set -e  # Exit on error

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting script...${NC}"

# Your code here

echo -e "${GREEN}✅ Done!${NC}"
```

---

## 🔧 Maintenance

**Keep scripts updated**:
- Test scripts after major refactoring
- Update paths if project structure changes
- Document breaking changes in script comments

---

**Last Updated**: 8 Desember 2024
