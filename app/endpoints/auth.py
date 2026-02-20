from fastapi import APIRouter, Depends, Response, Form
from core.config import security, config



# Создаем роутер с префиксом и тегами
router = APIRouter(
    prefix="/auth",           # Все пути будут начинаться с /auth
    tags=["authentication"],  # Для группировки в Swagger UI
    responses={404: {"description": "Not found"}}  # Общие ответы
)

VALID_ROLES = ("student", "analyst", "admissions")


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


# @router.post("/register", response_model=UserResponse, status_code=status.HTTP_201_CREATED)
# async def register(
#     user_data: UserCreate,
#     db: Session = Depends(get_db)
# ):
#     """Регистрация нового пользователя"""
#     # Создаем зависимости
#     user_repo = UserRepository(db)
#     auth_service = AuthService(user_repo)
    
#     # Вызываем бизнес-логику
#     new_user = await auth_service.register(user_data)
#     return new_user

# @router.post("/login", response_model=Token)
# async def login(
#     email: str = Form(...),
#     password: str = Form(...),
#     db: Session = Depends(get_db)
# ):
#     """Вход в систему"""
#     user_repo = UserRepository(db)
#     auth_service = AuthService(user_repo)
    
#     tokens = await auth_service.authenticate(email, password)
#     return tokens

# @router.post("/logout")
# async def logout(
#     token: str = Depends(oauth2_scheme),
#     db: Session = Depends(get_db)
# ):
#     """Выход из системы"""
#     # Логика выхода
#     return {"message": "Successfully logged out"}