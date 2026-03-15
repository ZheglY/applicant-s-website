# Unik Admissions Portal

Веб‑приложение для работы приёмной комиссии и абитуриентов: новости, аналитика, регистрация, личный кабинет и список поступающих.

## Возможности
- Новости приёмной комиссии 
- Аналитика кампании 
- Регистрация абитуриентов и личный кабинет.
- Общий список поступающих и место в рейтинге 

## Стек
- FastAPI + Jinja2
- SQLAlchemy (async) + PostgreSQL
- AuthX (JWT в HTTP‑only cookie)

## Быстрый старт через Docker (рекомендуется)

1. Проверьте, что установлен Docker Desktop.
2. Запустите сервисы:
```
docker compose up --build
```
3. Откройте: `http://localhost:8000`

### Дефолтные учётные записи
- Приёмная комиссия: `admin@unik.edu` / `admin`
- Аналитик: `prepod@unik.edu` / `123456`

### Порты
- Приложение: `8000`
- Postgres: `5433` на хосте → `5432` в контейнере

## Локальный запуск без Docker

1. Создайте виртуальное окружение и установите зависимости:
```
python -m venv venv
venv\Scripts\activate
pip install -r requirements.txt
```
2. Подготовьте `.env` (пример в `.env.template`):
```
DATABASE_URL=postgresql+asyncpg://admin:123@localhost:5433/unik
SECRET_KEY=change_me
DEFAULT_ADMISSIONS_LOGIN=admin@unik.edu
DEFAULT_ADMISSIONS_PASSWORD=admin
DEFAULT_ANALYST_LOGIN=prepod@unik.edu
DEFAULT_ANALYST_PASSWORD=123456
```
3. Убедитесь, что Postgres запущен, затем стартуйте приложение:
```
python app/main.py
```
4. Откройте: `http://127.0.0.1:8000`

## Миграции (Alembic)

Базовая миграция уже добавлена. Для применения:
```
alembic upgrade head
```

Создание новой миграции:
```
alembic revision --autogenerate -m "your message"
```


## Структура проекта
```
app/
├── core/                # конфиг, зависимости, роли, шаблоны
├── db/                  # движок, модели, инициализация
├── endpoints/           # роуты FastAPI
├── repositories/        # доступ к данным (analitics_repository.py, user_repository.py)
├── schemas/             # Pydantic‑схемы
├── services/            # бизнес‑логика
├── static/              # CSS/JS
├── templates/           # Jinja2‑шаблоны страниц
└── main.py              # точка входа
```

## Переменные окружения
- `DATABASE_URL` — строка подключения к PostgreSQL.
- `SECRET_KEY` — ключ подписи JWT.
- `DEFAULT_ADMISSIONS_LOGIN` / `DEFAULT_ADMISSIONS_PASSWORD` — учётка приёмной комиссии.
- `DEFAULT_ANALYST_LOGIN` / `DEFAULT_ANALYST_PASSWORD` — учётка аналитика.


