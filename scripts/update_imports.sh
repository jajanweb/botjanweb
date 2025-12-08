#!/bin/bash
set -e

echo "🔄 UPDATING ALL IMPORT PATHS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Function to update imports in a directory
update_imports_in_dir() {
    local dir=$1
    echo "📝 Updating: $dir"
    
    find "$dir" -name "*.go" -type f -exec sed -i \
        -e 's|github.com/exernia/botjanweb/internal/usecase/qris|github.com/exernia/botjanweb/internal/application/service/qris|g' \
        -e 's|github.com/exernia/botjanweb/internal/usecase/payment|github.com/exernia/botjanweb/internal/application/service/payment|g' \
        -e 's|github.com/exernia/botjanweb/internal/usecase/family|github.com/exernia/botjanweb/internal/application/service/family|g' \
        -e 's|github.com/exernia/botjanweb/internal/usecase/account|github.com/exernia/botjanweb/internal/application/service/account|g' \
        -e 's|github.com/exernia/botjanweb/internal/usecase"|github.com/exernia/botjanweb/internal/application/service"|g' \
        -e 's|github.com/exernia/botjanweb/internal/repository/memory|github.com/exernia/botjanweb/internal/infrastructure/persistence/memory|g' \
        -e 's|github.com/exernia/botjanweb/internal/repository/postgres|github.com/exernia/botjanweb/internal/infrastructure/persistence/postgres|g' \
        -e 's|github.com/exernia/botjanweb/internal/repository/sheets|github.com/exernia/botjanweb/internal/infrastructure/persistence/sheets|g' \
        -e 's|github.com/exernia/botjanweb/internal/infrastructure/whatsapp|github.com/exernia/botjanweb/internal/infrastructure/messaging/whatsapp|g' \
        -e 's|github.com/exernia/botjanweb/internal/infrastructure/webhook|github.com/exernia/botjanweb/internal/infrastructure/messaging/webhook|g' \
        -e 's|github.com/exernia/botjanweb/internal/infrastructure/qris|github.com/exernia/botjanweb/internal/infrastructure/external/qris|g' \
        -e 's|github.com/exernia/botjanweb/internal/controller/bot|github.com/exernia/botjanweb/presentation/handler/bot|g' \
        -e 's|github.com/exernia/botjanweb/internal/controller/http|github.com/exernia/botjanweb/presentation/handler/http|g' \
        {} \;
}

# Update imports in all new directories
update_imports_in_dir "internal/application"
update_imports_in_dir "internal/infrastructure"
update_imports_in_dir "presentation/handler"
update_imports_in_dir "internal/bootstrap"
update_imports_in_dir "cmd"

echo ""
echo "✅ All imports updated!"
