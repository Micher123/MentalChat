# 🔧 Решение проблемы с запуском

## Текущая проблема:
```
❌ Backend не запущен (ECONNREFUSED 127.0.0.1:8080)
❌ PostgreSQL не настроен
```

## ✅ Быстрое решение (выберите один вариант):

### Вариант A: Docker (2 минуты) ⭐ РЕКОМЕНДУЕТСЯ

```bash
# 1. Запустить PostgreSQL
docker run --name mentalchat-db \
  -e POSTGRES_DB=mentalchat \
  -e POSTGRES_USER=mentalchat_user \
  -e POSTGRES_PASSWORD=mentalchat_pass \
  -p 5432:5432 \
  -d postgres:14

# 2. Обновить .env.json (пароль: mentalchat_pass)

# 3. Запустить backend
cd Backend
./mentalchat

# 4. Проверить
curl http://localhost:8080/api/v1/config/trial
```

---

### Вариант B: Установить PostgreSQL (10 минут)

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install postgresql postgresql-contrib

# Запустить
sudo systemctl start postgresql

# Создать БД
sudo -u postgres psql << EOF
CREATE DATABASE mentalchat;
CREATE USER mentalchat_user WITH PASSWORD 'mentalchat_pass';
GRANT ALL PRIVILEGES ON DATABASE mentalchat TO mentalchat_user;
EOF

# Обновить .env.json
# Запустить backend
```

---

### Вариант C: Тест без backend (прямо сейчас!) 🎯

**Frontend уже работает! Можно проверить fingerprinting:**

1. Откройте http://localhost:3000/registration
2. Откройте консоль (F12)
3. Вы увидите: `Fingerprint generated: 9ca3baa88b671ab6...`
4. ✅ **Fingerprinting работает!**

Ошибка 500 возникает только при попытке сохранить в БД.

**Для полной проверки:**
- Откройте `test_fingerprint.html` в браузере
- Нажмите "🚀 Сгенерировать Fingerprint"
- Проверьте стабильность (кнопка 2-3 раза)

---

## Что уже работает:

✅ **Frontend собран** - http://localhost:3000  
✅ **Fingerprinting генерируется** - см. консоль браузера  
✅ **WelcomeModal исправлен** - плавная анимация  
✅ **Регистрация исправлена** - mentalState не требуется  

## Что не работает:

❌ **Backend не запущен** - нужна БД  
❌ **Регистрация не завершается** - ошибка 500 при сохранении  

---

## Рекомендация:

**Сейчас:** Протестируйте fingerprinting через `test_fingerprint.html`

**Потом:** Поднимите PostgreSQL через Docker (Вариант A)

Это займет 2 минуты и всё заработает полностью! 🚀

---

## Команды для копирования:

```bash
# Docker (самый быстрый способ)
docker run --name mentalchat-db -e POSTGRES_DB=mentalchat -e POSTGRES_USER=mentalchat_user -e POSTGRES_PASSWORD=mentalchat_pass -p 5432:5432 -d postgres:14

# Проверка что БД работает
docker ps | grep mentalchat

# Обновить .env.json
nano .env.json
# Изменить password на "mentalchat_pass"

# Запустить backend
cd Backend && ./mentalchat

# Проверить backend
curl http://localhost:8080/api/v1/config/trial
```
