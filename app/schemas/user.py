import re
from datetime import date
from typing import Optional

from pydantic import BaseModel, field_validator, model_validator


def _email_regex(v: str) -> str:
    pattern = r"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$"
    if not re.match(pattern, v):
        raise ValueError("Некорректный формат email")
    return v.lower()


class ApplicantRegisterSchema(BaseModel):
    """Схема регистрации абитуриента с валидацией."""

    fullname: str
    password: str
    password_confirm: str
    birthdate: date
    phone: str
    email: str
    telegram: Optional[str] = None
    school: str
    achievements: Optional[str] = None
    priorities: list[str]  # до 3 направлений
    agreement: bool = True
    ege_scores: Optional[dict[str, int]] = None  # {"Русский язык": 90, ...}

    @field_validator("ege_scores", mode="before")
    @classmethod
    def parse_ege_scores(cls, v):
        if v is None or not isinstance(v, dict):
            return {}
        result = {}
        for k, val in v.items():
            if val and str(val).strip():
                try:
                    result[k.strip()] = int(val)
                except (ValueError, TypeError):
                    pass
        return result if result else None

    @field_validator("email")
    @classmethod
    def validate_email(cls, v: str) -> str:
        return _email_regex(v)

    @field_validator("fullname")
    @classmethod
    def fullname_not_empty(cls, v: str) -> str:
        parts = v.strip().split()
        if len(parts) < 2:
            raise ValueError("Укажите ФИО (имя и фамилию)")
        return v.strip()

    @field_validator("password")
    @classmethod
    def password_min_length(cls, v: str) -> str:
        if len(v) < 6:
            raise ValueError("Пароль должен быть не менее 6 символов")
        return v

    @model_validator(mode="after")
    def passwords_match(self):
        if self.password != self.password_confirm:
            raise ValueError("Пароли не совпадают")
        return self

    @field_validator("priorities")
    @classmethod
    def priorities_not_empty(cls, v: list[str]) -> list[str]:
        clean = [p.strip() for p in v if p and p.strip()]
        if not clean:
            raise ValueError("Выберите хотя бы одно направление")
        if len(clean) > 3:
            raise ValueError("Максимум 3 направления")
        return clean[:3]

    @field_validator("agreement")
    @classmethod
    def agreement_required(cls, v: bool) -> bool:
        if not v:
            raise ValueError("Необходимо согласие на обработку персональных данных")
        return v

    def split_fullname(self) -> tuple[str, str, str | None]:
        """Разбивает ФИО на фамилию, имя, отчество."""
        parts = self.fullname.strip().split(maxsplit=2)
        last_name = parts[0]
        first_name = parts[1] if len(parts) > 1 else ""
        middle_name = parts[2] if len(parts) > 2 else None
        return last_name, first_name, middle_name


ApplicantSchema = ApplicantRegisterSchema  # алиас для совместимости


class UserLoginSchema(BaseModel):
    """Проверка логина и пароля пользователя."""

    username: str
    password: str
