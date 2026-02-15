from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles
from fastapi.templating import Jinja2Templates
# from app.core.config import settings
# from app.core.logging_config import setup_logging
from app.endpoints import auth_router, users_router  # Импортируем роутеры

# Настройка логирования
# setup_logging()

# Создаем приложение
app = FastAPI(
    # title=settings.PROJECT_NAME,
    # version=settings.VERSION,
    # debug=settings.DEBUG,
    docs_url="/api/docs",  # Swagger UI
    redoc_url="/api/redoc"  # ReDoc
)

# Подключаем статические файлы (если есть)
# app.mount("/static", StaticFiles(directory="app/static"), name="static")

# ПОДКЛЮЧАЕМ РОУТЕРЫ
app.include_router(auth_router)      # Все пути /auth/*
app.include_router(users_router)     # Все пути /users/*

# Шаблоны для HTML
# templates = Jinja2Templates(directory="app/templates")

@app.get("/index")
async def index():
    return {'message': "wkdwkwopfwelfwF"}



