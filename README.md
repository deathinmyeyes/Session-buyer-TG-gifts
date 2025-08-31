# Gift Buyer - автоматическая покупка Telegram подарков

[![Language: Russian](https://img.shields.io/badge/Language-Русский-blue)](#русский) [![Language: English](https://img.shields.io/badge/Language-English-green)](#english) [![License: CC BY-NC-ND 4.0](https://img.shields.io/badge/License-CC%20BY--NC--ND%204.0-lightgrey.svg)](https://creativecommons.org/licenses/by-nc-nd/4.0/) [![Telegram](https://img.shields.io/badge/Telegram-@earnfame-blue?logo=telegram)](https://t.me/earnfame)

## 📖 Описание

**Gift Buyer** — автоматизированная система для покупки Star Gifts в Telegram. Программа непрерывно мониторит доступные подарки, проверяет их соответствие заданным критериям и автоматически покупает подходящие варианты.

### ⚡ Основные возможности

- **🎯 Умная фильтрация** — настраиваемые критерии по цене, лимитам и количеству
- **👥 Гибкие получатели** — поддержка username'ов для пользователей и каналов
- **⚡ Высокая скорость** — параллельная обработка и настраиваемый тикер мониторинга
- **📱 Уведомления** — интеграция с Telegram Bot для уведомлений о статусе покупок
- **🔄 Автоматический реконнект** — устойчивость к сбоям API с автоматическим переподключением
- **💾 Кэширование** — сохранение состояния между перезапусками

## 🚀 Быстрый старт

### 1. Установка и сборка

```bash
git clone <repository-url>
cd Session-buyer-TG-gifts
go build -o Session-buyer-TG-gifts.exe cmd/main.go
```

### 2. Настройка конфигурации

```bash
cp internal/config/config_example.json internal/config/config.json
```

Отредактируйте `config.json` с вашими данными (см. [Конфигурация](#️-конфигурация)).

### 3. Запуск

```bash
./Session-buyer-TG-gifts
```

## 🧪 Тестирование

Для тестирования работы Gift Buyer выполните следующие шаги:

1. **Включите тестовый режим** — в файле `config.json` установите:
   ```json
   "gift_param": {
       ...
       "test_mode": true,
       ...
   }
   ```
2. **Отключите лимитированность подарков** — установите:
   ```json
   "gift_param": {
       ...
       "limited_status": false,
       ...
   }
   ```
3. **Удалите файл кэша** — перед запуском теста удалите файл `cache.json` (если он существует) для чистого старта:
   ```bash
   rm internal/config/cache.json
   ```

Теперь можно запускать приложение для тестирования без ограничений и с чистым состоянием.

## ⚙️ Конфигурация

### 🔧 Telegram настройки

```json
{
    "tg_settings": {
        "app_id": 12345678,
        "api_hash": "ваш_api_hash",
        "phone": "+1234567890",
        "password": "пароль_2fa",
        "tg_bot_key": "токен_бота",
        "datacenter": 4,
        "notification_chat_id": 123456789
    }
}
```

- **`app_id`** и **`api_hash`** — получите на [my.telegram.org](https://my.telegram.org)
- **`phone`** — номер телефона в международном формате
- **`password`** — пароль 2FA (оставьте `""` если отключена)
- **`tg_bot_key`** — токен бота для уведомлений (опционально)
- **`datacenter`** — датацентр Telegram (0=авто, 1-8=конкретный ДЦ). Рекомендуется 4 если DC2 лагает
- **`notification_chat_id`** — ваш User ID для уведомлений

**💡 Датацентры:** DC1/DC3 (Майами), DC2/DC4 (Амстердам), DC5 (Сингапур), DC8 (Франкфурт). DC4 рекомендуется для стабильности.

### 🎯 Критерии покупки

```json
{
    "criterias": [
        {
            "min_price": 10,
            "max_price": 100,
            "total_supply": 100000000,
            "count": 10,
            "receiver_type": [1]
        }
    ]
}
```

- **`min_price`/`max_price`** — ценовой диапазон в звездах
- **`total_supply`** — максимальный тираж подарка
- **`count`** — количество подарков для покупки
- **`receiver_type`** — типы получателей: `[0]` - себе, `[1]` - пользователям, `[2]` - каналам

### 👤 Получатели подарков

```json
{
    "receiver": {
        "user_receiver_id": ["durovs_dog", "username2"],
        "channel_receiver_id": ["durov", "telegram"]
    }
}
```

**Важно:** Используйте только **username'ы** (без @), НЕ числовые ID.

### 🚀 Параметры производительности

```json
{
    "ticker": 2.0,
    "retry_count": 5,
    "retry_delay": 2.5,
    "concurrency_gift_count": 10,
    "concurrent_operations": 300,
    "rpc_rate_limit": 30
}
```

- **`ticker`** — интервал мониторинга в секундах
- **`retry_count`** — количество попыток при ошибках
- **`rpc_rate_limit`** — лимит RPC запросов в секунду

### ⚙️ Параметры подарков

```json
{
    "gift_param": {
        "total_star_cap": 1000000000000,
        "limited_status": true,
        "release_by": false,
        "test_mode": false
    }
}
```

- **`total_star_cap`** — максимальное количество звезд для покупки всех подарков
- **`limited_status`** — покупать только ограниченные (true) или неограниченные (false) подарки
- **`release_by`** — проверять наличие информации о релизере подарка
- **`test_mode`** — тестовый режим, отключает проверки лимитов

### 🔒 Глобальные ограничения

```json
{
    "max_buy_count": 100
}
```

### 📋 Полный пример конфигурации

См. файл [`internal/config/config_example.json`](internal/config/config_example.json)

## 📄 Лицензия

Этот проект распространяется под лицензией **Creative Commons Attribution-NonCommercial-NoDerivatives 4.0 International (CC BY-NC-ND 4.0)**.

### Что разрешено:
✅ Использование для личных целей  
✅ Изучение кода  
✅ Распространение ссылки на проект  

### Что запрещено:
❌ Коммерческое использование  
❌ Перепродажа или продажа  
❌ Создание форков для распространения  
❌ Производные работы  

⚖️ **Юридическая защита:** Лицензия имеет международную юридическую силу.

Полный текст: [CC BY-NC-ND 4.0](https://creativecommons.org/licenses/by-nc-nd/4.0/)

---

**Для коммерческого использования** свяжитесь с автором.

---

## English

## 📖 Description

**Gift Buyer** is an automated system for purchasing Star Gifts in Telegram. The program continuously monitors available gifts, validates them against configured criteria, and automatically purchases eligible options.

### ⚡ Key Features

- **🎯 Smart Filtering** — configurable criteria by price, limits, and quantity
- **👥 Flexible Recipients** — support for usernames for users and channels
- **⚡ High Speed** — parallel processing and configurable monitoring ticker
- **📱 Notifications** — Telegram Bot integration for purchase status notifications
- **🔄 Auto Reconnect** — resilience to API failures with automatic reconnection
- **💾 Caching** — state persistence between restarts

## 🚀 Quick Start

### 1. Installation and Build

```bash
git clone <repository-url>
cd Session-buyer-TG-gifts
go build -o Session-buyer-TG-gifts.exe cmd/main.go
```

### 2. Configuration Setup

```bash
cp internal/config/config_example.json internal/config/config.json
```

Edit `config.json` with your data (see [Configuration](#️-configuration-1)).

### 3. Launch

```bash
./Session-buyer-TG-gifts
```

## 🧪 Тестирование

Для тестирования работы Gift Buyer выполните следующие шаги:

1. **Включите тестовый режим** — в файле `config.json` установите:
   ```json
   "gift_param": {
       ...
       "test_mode": true,
       ...
   }
   ```
2. **Отключите лимитированность подарков** — установите:
   ```json
   "gift_param": {
       ...
       "limited_status": false,
       ...
   }
   ```
3. **Удалите файл кэша** — перед запуском теста удалите файл `cache.json` (если он существует) для чистого старта:
   ```bash
   rm internal/config/cache.json
   ```

Теперь можно запускать приложение для тестирования без ограничений и с чистым состоянием.

## ⚙️ Configuration

### 🔧 Telegram Settings

```json
{
    "tg_settings": {
        "app_id": 12345678,
        "api_hash": "your_api_hash",
        "phone": "+1234567890",
        "password": "2fa_password",
        "tg_bot_key": "bot_token",
        "notification_chat_id": 123456789
    }
}
```

- **`app_id`** and **`api_hash`** — get from [my.telegram.org](https://my.telegram.org)
- **`phone`** — phone number in international format
- **`password`** — 2FA password (leave `""` if disabled)
- **`tg_bot_key`** — bot token for notifications (optional)
- **`notification_chat_id`** — your User ID for notifications

### 🎯 Purchase Criteria

```json
{
    "criterias": [
        {
            "min_price": 10,
            "max_price": 100,
            "total_supply": 100000000,
            "count": 10,
            "receiver_type": [1]
        }
    ]
}
```

- **`min_price`/`max_price`** — price range in stars
- **`total_supply`** — maximum gift supply
- **`count`** — number of gifts to purchase
- **`receiver_type`** — recipient types: `[0]` - self, `[1]` - users, `[2]` - channels

### 👤 Gift Recipients

```json
{
    "receiver": {
        "user_receiver_id": ["durovs_dog", "username2"],
        "channel_receiver_id": ["durov", "telegram"]
    }
}
```

**Important:** Use only **usernames** (without @), NOT numeric IDs.

### 🚀 Performance Parameters

```json
{
    "ticker": 2.0,
    "retry_count": 5,
    "retry_delay": 2.5,
    "concurrency_gift_count": 10,
    "concurrent_operations": 300,
    "rpc_rate_limit": 30
}
```

- **`ticker`** — monitoring interval in seconds
- **`retry_count`** — number of retry attempts on errors
- **`rpc_rate_limit`** — RPC requests limit per second

### ⚙️ Gift Parameters

```json
{
    "gift_param": {
        "total_star_cap": 1000000000000,
        "limited_status": true,
        "release_by": false,
        "test_mode": false
    }
}
```

- **`total_star_cap`** — maximum stars for purchasing all gifts
- **`limited_status`** — buy only limited (true) or unlimited (false) gifts
- **`release_by`** — check for gift releaser information
- **`test_mode`** — test mode, disables limit validations

### 🔒 Global Limits

```json
{
    "max_buy_count": 100
}
```

### 📋 Full Configuration Example

See file [`internal/config/config_example.json`](internal/config/config_example.json)

## 📄 License

This project is distributed under **Creative Commons Attribution-NonCommercial-NoDerivatives 4.0 International (CC BY-NC-ND 4.0)** license.

### What's allowed:
✅ Personal use  
✅ Code study  
✅ Project link sharing  

### What's prohibited:
❌ Commercial use  
❌ Reselling or selling  
❌ Creating forks for distribution  
❌ Derivative works  

⚖️ **Legal Protection:** License has international legal force.

Full text: [CC BY-NC-ND 4.0](https://creativecommons.org/licenses/by-nc-nd/4.0/)

---

**For commercial use** contact the author.
