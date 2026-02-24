from datetime import datetime

from fastapi import APIRouter, Depends, HTTPException, Request, Response, Form
from fastapi.responses import HTMLResponse, RedirectResponse
from fastapi.templating import Jinja2Templates
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from core.config import security, config
from core.templates import templates
from db.engine import get_session, SessionDep
from schemas.user import ApplicantRegisterSchema
from services.auth_service import register_user


# Создаем роутер с префиксом и тегами
router = APIRouter(
    prefix="/auth",           # Все пути будут начинаться с /auth
    tags=["authentication"],  # Для группировки в Swagger UI
    responses={404: {"description": "Not found"}}  # Общие ответы
)

VALID_ROLES = ("student", "analyst", "admissions")


@router.get("/register", response_class=HTMLResponse)
async def register_page(request: Request):
    """Страница регистрации абитуриента (шаблон regist.html)."""
    return templates.TemplateResponse("regist.html", {"request": request})


@router.get("/enter", response_class=HTMLResponse)
async def login_page(request: Request):
    """Страница входа."""
    return templates.TemplateResponse("enter.html", {"request": request})


@router.get("/login", response_class=HTMLResponse)
async def login_redirect():
    """Редирект на страницу входа (GET /auth/login → /auth/enter)."""
    return RedirectResponse(url="/auth/enter", status_code=302)


@router.post("/register")
async def register(
    data: ApplicantRegisterSchema,
    session: SessionDep,
):
    """
    Регистрация нового абитуриента в БД с валидацией Pydantic.
    """
    
    applicant = await register_user(session, data)
    return {
    "message": "Регистрация успешна",
    "id": applicant.id,
    "email": applicant.email,
    }


@router.post("/login")
def login(response: Response, role: str = Form("student", description="student | analyst | admissions")):
    """
    Вход в систему. role берётся из БД (студент/абитуриент, аналитик, приёмная комиссия).
    Временный параметр role для тестирования разных ролей.
    Роли: student, analyst, admissions
    """
    if role not in VALID_ROLES:
        role = "student"
    # TODO: uid и role получать из БД после проверки логина/пароля
    token = security.create_access_token(
        uid="12345",
        data={"role": role}
    )
    response.set_cookie(config.JWT_ACCESS_COOKIE_NAME, token)
    return {"message": "Успешный вход", "role": role}



# @router.post("/logout")
# async def logout(
#     token: str = Depends(oauth2_scheme),
#     db: Session = Depends(get_db)
# ):
#     """Выход из системы"""
#     # Логика выхода
#     return {"message": "Successfully logged out"}