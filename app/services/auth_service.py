from datetime import datetime
from fastapi import HTTPException

from db.models import Applicant
from repositories.user_repository import get_by_id, create_applicant, get_direction_by_name
from utils.hash_password import _hash_password


async def register_user(session, data):
    """
    Регистрация абитуриента (используется в эндпоинте регистрации)
    1. Проверяется уникальность почты т.к это логин пользователя
    2. Проверка корретного ввода баллов егэ
    3. Хеширование пароля и создания пользователя
    """

    existing = await get_by_id(session, data.email)
    if existing:
        raise HTTPException(
            status_code=400,
            detail="Пользователь с таким email уже зарегистрирован"
        )

    faculty_direction = await get_direction_by_name(session, data.priorities[0])
    if not faculty_direction:
        raise HTTPException(
            status_code=400,
            detail="Направление не найдено"
        )

    if not data.ege_scores:
        raise HTTPException(
            status_code=400,
            detail="Не указаны баллы егэ"
        )

    total_score = sum(data.ege_scores.values())
    last_name, first_name, middle_name = data.split_fullname()
    

    applicant = Applicant(
    last_name=last_name,
    first_name=first_name,
    middle_name=middle_name,
    email=data.email,
    login=data.email,
    password_hash=_hash_password(data.password),
    phone=data.phone or None,
    telegram=data.telegram or None,
    birth_date=datetime.combine(data.birthdate, datetime.min.time()),
    total_score=total_score,
    role="student",
    sex=True,
    achievements=data.achievements or None,
    school=data.school,
    status_code=0,
    faculty_direction_id=faculty_direction.id,
    )

    return await create_applicant(session, applicant=applicant)