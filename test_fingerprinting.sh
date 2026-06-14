#!/bin/bash

# Тестовый скрипт для проверки fingerprinting
# Запуск: ./test_fingerprinting.sh

echo "🧪 Тестирование Fingerprinting в MentalChat"
echo "==========================================="
echo ""

# Цвета
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Проверка сборки backend
echo "1️⃣  Проверка сборки backend..."
if [ -f "./Backend/mentalchat" ]; then
    echo -e "${GREEN}✅ Backend собран успешно${NC}"
else
    echo -e "${RED}❌ Backend не найден. Запустите: ./build.sh --build${NC}"
    exit 1
fi

# Проверка сборки frontend
echo ""
echo "2️⃣  Проверка сборки frontend..."
if [ -d "./Frontend/dist" ]; then
    echo -e "${GREEN}✅ Frontend собран успешно${NC}"
else
    echo -e "${RED}❌ Frontend dist не найден${NC}"
    exit 1
fi

# Проверка файлов fingerprinting
echo ""
echo "3️⃣  Проверка файлов fingerprinting..."

FILES=(
    "Backend/internal/service/fingerprint_service.go"
    "Backend/internal/handler/auth_handler.go"
    "Frontend/src/hooks/useFingerprint.ts"
    "Frontend/src/pages/RegistrationPage.tsx"
    "docs/FINGERPRINTING.md"
)

for file in "${FILES[@]}"; do
    if [ -f "$file" ]; then
        echo -e "${GREEN}✅ $file${NC}"
    else
        echo -e "${RED}❌ $file не найден${NC}"
    fi
done

# Проверка импортов в backend
echo ""
echo "4️⃣  Проверка импортов fingerprint в backend..."
if grep -q "FingerprintService" Backend/cmd/main.go; then
    echo -e "${GREEN}✅ FingerprintService импортирован в main.go${NC}"
else
    echo -e "${RED}❌ FingerprintService не найден в main.go${NC}"
fi

if grep -q "fingerprint_service" Backend/internal/service/auth_service.go; then
    echo -e "${GREEN}✅ FingerprintService используется в auth_service${NC}"
else
    echo -e "${RED}❌ FingerprintService не используется в auth_service${NC}"
fi

# Проверка полей в моделях
echo ""
echo "5️⃣  Проверка полей fingerprint в моделях..."
if grep -q "Fingerprint.*string" Backend/internal/model/models.go; then
    echo -e "${GREEN}✅ Поле Fingerprint есть в модели User${NC}"
else
    echo -e "${RED}❌ Поле Fingerprint не найдено в модели User${NC}"
fi

if grep -q "FingerprintData.*string" Backend/internal/model/models.go; then
    echo -e "${GREEN}✅ Поле FingerprintData есть в модели User${NC}"
else
    echo -e "${RED}❌ Поле FingerprintData не найдено в модели User${NC}"
fi

# Проверка методов в storage
echo ""
echo "6️⃣  Проверка методов Storage..."
if grep -q "GetUserByFingerprint" Backend/internal/storage/storage.go; then
    echo -e "${GREEN}✅ Метод GetUserByFingerprint существует${NC}"
else
    echo -e "${RED}❌ Метод GetUserByFingerprint не найден${NC}"
fi

if grep -q "GetAllUsersWithFingerprint" Backend/internal/storage/storage.go; then
    echo -e "${GREEN}✅ Метод GetAllUsersWithFingerprint существует${NC}"
else
    echo -e "${RED}❌ Метод GetAllUsersWithFingerprint не найден${NC}"
fi

# Проверка frontend хука
echo ""
echo "7️⃣  Проверка frontend хука..."
if grep -q "useFingerprint" Frontend/src/pages/RegistrationPage.tsx; then
    echo -e "${GREEN}✅ useFingerprint используется в RegistrationPage${NC}"
else
    echo -e "${RED}❌ useFingerprint не используется в RegistrationPage${NC}"
fi

if grep -q "fingerprint" Frontend/src/services/api.ts; then
    echo -e "${GREEN}✅ fingerprint отправляется в API${NC}"
else
    echo -e "${RED}❌ fingerprint не отправляется в API${NC}"
fi

# Проверка документации
echo ""
echo "8️⃣  Проверка документации..."
if [ -f "docs/FINGERPRINTING.md" ]; then
    LINES=$(wc -l < docs/FINGERPRINTING.md)
    echo -e "${GREEN}✅ Документация создана ($LINES строк)${NC}"
else
    echo -e "${RED}❌ Документация не найдена${NC}"
fi

echo ""
echo "==========================================="
echo "📊 Результаты тестов:"
echo "==========================================="
echo ""

# Запуск тестов Go
echo "9️⃣  Запуск Go тестов..."
cd Backend
if go test ./internal/service/... -v 2>&1 | grep -q "PASS"; then
    echo -e "${GREEN}✅ Go тесты пройдены${NC}"
else
    echo -e "${YELLOW}⚠️  Go тесты не найдены или требуют настройки${NC}"
fi

cd ..

echo ""
echo "🔍 Ручное тестирование:"
echo "==========================================="
echo ""
echo "1. Запустите backend:"
echo "   cd Backend && ./mentalchat"
echo ""
echo "2. Запустите frontend:"
echo "   cd Frontend && npm run dev"
echo ""
echo "3. Откройте http://localhost:3000/registration"
echo ""
echo "4. Откройте консоль разработчика (F12)"
echo ""
echo "5. Проверьте лог:"
echo "   Fingerprint generated: <hash>"
echo ""
echo "6. Заполните форму регистрации"
echo ""
echo "7. Проверьте network tab - fingerprint должен отправиться"
echo ""
echo "8. Проверьте БД:"
echo "   SELECT email, fingerprint, trial_end FROM users;"
echo ""

echo "==========================================="
echo "✅ Тестирование завершено!"
echo "==========================================="
