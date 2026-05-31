from contextlib import asynccontextmanager
from pathlib import Path

import uvicorn
from fastapi import FastAPI, Request
from fastapi.responses import RedirectResponse
from fastapi.staticfiles import StaticFiles
from fastapi.templating import Jinja2Templates

from app.db.engine import setup_database, seed_directions, new_session
from app.endpoints import auth_router, users_router
from app.services.auth_service import ensure_staff_users

_BASE_DIR = Path(__file__).resolve().parent


@asynccontextmanager
async def lifespan(app: FastAPI):
    await setup_database()
    await seed_directions()
    async with new_session() as session:
        await ensure_staff_users(session)
    yield


# Создаем приложение
app = FastAPI(
    lifespan=lifespan,
    docs_url="/api/docs",
    redoc_url="/api/redoc",
)

templates = Jinja2Templates(directory=str(_BASE_DIR / "templates"))

# Подключаем статические файлы
app.mount("/static", StaticFiles(directory=str(_BASE_DIR / "static")), name="static")


@app.get("/", include_in_schema=False)
async def root_redirect():
    return RedirectResponse(url="/users/news", status_code=302)


# ПОДКЛЮЧАЕМ РОУТЕРЫ
app.include_router(auth_router)      # Все пути /auth/*
app.include_router(users_router)     # Все пути /users/*


if __name__ == "__main__":
    uvicorn.run("app.main:app", reload=True)
