#!/bin/bash

# Быстрый тестовый запуск с SQLite

echo "🚀 MentalChat Quick Test (SQLite)"
echo "=================================="

cd Backend

# Создаем временную конфигурацию для SQLite
cat > .env.test.json << 'EOF'
{
  "server": {
    "host": "localhost",
    "port": 8081,
    "debug": true
  },
  "database": {
    "type": "sqlite",
    "path": "./test.db"
  },
  "ai": {
    "chad_api_url": "https://ask.chadgpt.ru/api/public",
    "models": {
      "free": "gpt-4o-mini",
      "pro": "gpt-4o",
      "ultra": "gpt-4-turbo"
    },
    "yandex_speechkit": {
      "enabled": false
    }
  },
  "payment": {
    "provider": "yoomoney",
    "yoomoney_shop_id": "test",
    "yoomoney_secret": "test",
    "yoomoney_scid": "test",
    "prices": {
      "pro_monthly": 499,
      "pro_yearly": 4990,
      "ultra_monthly": 999,
      "ultra_yearly": 9990
    }
  },
  "email": {
    "smtp_host": "smtp.example.com",
    "smtp_port": 587,
    "smtp_user": "test@example.com",
    "smtp_password": "test",
    "from_email": "test@example.com",
    "from_name": "MentalChat Test"
  },
  "security": {
    "jwt_secret": "test_secret_key_for_testing_only",
    "jwt_expiration_hours": 168,
    "rate_limit_requests": 100,
    "rate_limit_window_seconds": 60,
    "ddos_protection_enabled": false,
    "max_concurrent_requests_per_ip": 50
  },
  "storage": {
    "voice_storage_path": "./storage/voices",
    "avatar_storage_path": "./storage/avatars",
    "max_file_size_mb": 10
  },
  "app": {
    "frontend_url": "http://localhost:3000",
    "app_name": "MentalChat Test",
    "support_email": "test@mentalchat.com"
  }
}
EOF

echo ""
echo "📝 Инструкция:"
echo "=============="
echo ""
echo "1. Backend будет запущен на порту 8081"
echo "2. Frontend должен быть на порту 3000"
echo "3. Данные сохраняются в Backend/test.db"
echo ""
echo "⚠️  Для работы нужен SQLite драйвер:"
echo "   cd Backend && go get gorm.io/driver/sqlite"
echo ""
echo "🔧 Или настройте PostgreSQL:"
echo "   docker run --name mentalchat-db -e POSTGRES_PASSWORD=pass -p 5432:5432 -d postgres:14"
echo ""

# Пробуем запустить с SQLite
if [ -f ".env.test.json" ]; then
    echo "✅ Тестовая конфигурация создана"
    echo ""
    echo "Запуск backend..."
    cp .env.json .env.json.backup 2>/dev/null
    cp .env.test.json .env.json
    ./mentalchat &
    BACKEND_PID=$!
    echo "Backend PID: $BACKEND_PID"
    sleep 2
    
    # Проверка запуска
    if curl -s http://localhost:8081/api/v1/config/trial > /dev/null 2>&1; then
        echo "✅ Backend успешно запущен!"
    else
        echo "❌ Backend не запустился. Проверьте логи."
        kill $BACKEND_PID 2>/dev/null
        cp .env.json.backup .env.json 2>/dev/null
    fi
fi
