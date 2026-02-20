from pydantic import BaseModel


class ApplicantSchema(BaseModel):
    """
    Схема абитруента для валидации данных при регистрации пользователя на сайте
    """
    pass


class UserLoginSchema(BaseModel):
    """
    Проверка логина и пароля пользователя
    """
    pass

