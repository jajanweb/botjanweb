#!/bin/bash
set -e

echo "🚀 COMPREHENSIVE STRUCTURE MIGRATION"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Phase 4: Migrate usecase → application
echo "📦 Phase 4: usecase → application"
cp -r internal/usecase/qris internal/application/service/
cp -r internal/usecase/payment internal/application/service/
cp -r internal/usecase/family internal/application/service/
cp -r internal/usecase/account internal/application/service/
echo "✅ Copied all usecase services"

# Phase 5: Migrate repository → infrastructure/persistence
echo ""
echo "💾 Phase 5: repository → infrastructure/persistence"
cp -r internal/repository/memory internal/infrastructure/persistence/
cp -r internal/repository/postgres internal/infrastructure/persistence/
cp -r internal/repository/sheets internal/infrastructure/persistence/
echo "✅ Copied all repositories"

# Phase 6: Reorganize infrastructure
echo ""
echo "🔧 Phase 6: Reorganize infrastructure"
cp -r internal/infrastructure/whatsapp internal/infrastructure/messaging/
cp -r internal/infrastructure/webhook internal/infrastructure/messaging/
cp -r internal/infrastructure/qris internal/infrastructure/external/
echo "✅ Reorganized infrastructure"

# Phase 7: Migrate controller → presentation/handler
echo ""
echo "🎨 Phase 7: controller → presentation/handler"
cp -r internal/controller/bot internal/presentation/handler/
cp -r internal/controller/http internal/presentation/handler/
echo "✅ Copied controllers to presentation/handler"

echo ""
echo "✅ ALL FILES COPIED! Next: Update imports"
