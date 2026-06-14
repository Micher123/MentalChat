# 🚀 Быстрый старт для тестирования

## Проблема: PostgreSQL не настроен

Ошибка `role "n1" does not exist` означает, что база данных PostgreSQL не настроена.

## Решение 1: Настроить PostgreSQL (рекомендуется)

### 1. Установите PostgreSQL (если не установлен):

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install postgresql postgresql-contrib

# Запустить PostgreSQL
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

### 2. Создайте базу данных и пользователя:

```bash
# Войти в PostgreSQL
sudo -u postgres psql

# Выполнить SQL команды:
CREATE DATABASE mentalchat;
CREATE USER mentalchat_user WITH PASSWORD 'mentalchat_pass';
GRANT ALL PRIVILEGES ON DATABASE mentalchat TO mentalchat_user;
\q
```

### 3. Обновите .env.json:

```json
{
  "database": {
    "host": "localhost",
    "port": 5432,
    "name": "mentalchat",
    "user": "mentalchat_user",
    "password": "mentalchat_pass",
    "ssl_mode": "disable"
  }
}
```

### 4. Запустите backend:

```bash
cd Backend
./mentalchat
```

**Ожидаемый лог:**
```
INFO  Starting server address=localhost:8080
```

---

## Решение 2: SQLite для быстрого теста (без установки PostgreSQL)

### 1. Измените storage.go на SQLite:

```go
// Временно используйте SQLite для тестов
db, err := gorm.Open(sqlite.Open("mentalchat.db"), &gorm.Config{})
```

### 2. Установите драйвер SQLite:

```bash
cd Backend
go get gorm.io/driver/sqlite
```

### 3. Обновите .env.json:

```json
{
  "database": {
    "type": "sqlite",
    "path": "./mentalchat.db"
  }
}
```

---

## Решение 3: Docker с PostgreSQL (самый простой способ)

### 1. Запустите PostgreSQL в Docker:

```bash
docker run --name mentalchat-db \
  -e POSTGRES_DB=mentalchat \
  -e POSTGRES_USER=mentalchat_user \
  -e POSTGRES_PASSWORD=mentalchat_pass \
  -p 5432:5432 \
  -d postgres:14
```

### 2. Обновите .env.json:

```json
{
  "database": {
    "host": "localhost",
    "port": 5432,
    "name": "mentalchat",
    "user": "mentalchat_user",
    "password": "mentalchat_pass",
    "ssl_mode": "disable"
  }
}
```

### 3. Запустите backend:

```bash
cd Backend
./mentalchat
```

### 4. Остановите Docker когда закончите:

```bash
docker stop mentalchat-db
docker rm mentalchat-db
```

---

## Проверка работы

### 1. Backend запущен:

```bash
curl http://localhost:8080/api/v1/config/trial
```

**Ожидаемый ответ:**
```json
{
  "tiers": [
    {"tier": "free", ...},
    {"tier": "pro", ...},
    {"tier": "ultra", ...}
  ]
}
```

### 2. Frontend запущен:

Откройте http://localhost:3000

### 3. Регистрация:

1. Откройте http://localhost:3000/registration
2. Заполните форму
3. Нажмите "Регистрация"
4. Проверьте консоль - не должно быть ошибок 500

---

## Частые ошибки

### ❌ "role does not exist"

**Причина:** Пользователь БД не создан

**Решение:**
```bash
sudo -u postgres psql
CREATE USER mentalchat_user WITH PASSWORD 'mentalchat_pass';
```

### ❌ "database does not exist"

**Причина:** База данных не создана

**Решение:**
```bash
sudo -u postgres psql
CREATE DATABASE mentalchat;
```

### ❌ "permission denied"

**Причина:** Нет прав у пользователя

**Решение:**
```bash
sudo -u postgres psql
GRANT ALL PRIVILEGES ON DATABASE mentalchat TO mentalchat_user;
```

### ❌ "port already in use"

**Причина:** PostgreSQL уже запущен на порту 5432

**Решение:**
```bash
# Найти процесс
sudo lsof -i :5432

# Или использовать другой порт в .env.json
"port": 5433
```

---

## Тестирование fingerprinting без БД

Даже без работающей БД можно протестировать генерацию fingerprint:

1. Откройте http://localhost:3000/registration
2. Откройте консоль (F12)
3. Вы увидите: `Fingerprint generated: 9ca3baa88b671ab6...`
4. Это значит frontend работает правильно!

Ошибка 500 возникает только при попытке сохранить в БД.

---

## Рекомендация

**Используйте Docker** - это самый быстрый способ поднять PostgreSQL для тестов:

```bash
# Запустить
docker run --name mentalchat-db \
  -e POSTGRES_DB=mentalchat \
  -e POSTGRES_USER=mentalchat_user \
  -e POSTGRES_PASSWORD=mentalchat_pass \
  -p 5432:5432 \
  -d postgres:14

# Протестировать
cd MentalChat
./build.sh --run

# Остановить
docker stop mentalchat-db
```

Это займет 2 минуты! 🚀
