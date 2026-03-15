import re

from fastapi import HTTPException

from core.config import app_config
from db.models import Applicant, ApplicantPriority
from repositories.user_repository import get_by_email, get_by_login, get_direction_by_name
from utils.hash_password import hash_password, verify_password


def _parse_achievements(text: str | None) -> list[dict] | None:
    if not text:
        return None
    parts = [p.strip() for p in re.split(r"[;\n,]+", text) if p.strip()]
    if not parts:
        return None
    return [{"text": item, "points": 1} for item in parts]


async def register_user(session, data):
    existing = await get_by_email(session, data.email)
    if existing:
        raise HTTPException(
            status_code=400,
            detail="User with this email already exists",
        )

    if not data.ege_scores:
        raise HTTPException(
            status_code=400,
            detail="EGE scores are required",
        )

    total_score = sum(data.ege_scores.values())
    last_name, first_name, middle_name = data.split_fullname()

    applicant = Applicant(
        last_name=last_name,
        first_name=first_name,
        middle_name=middle_name,
        email=data.email,
        login=data.email,
        password_hash=hash_password(data.password),
        phone=data.phone or None,
        telegram=data.telegram or None,
        birth_date=data.birthdate,
        total_score=total_score,
        role="student",
        sex=True,
        achievements=_parse_achievements(data.achievements),
        school=data.school,
        region=None,
        ege_scores=data.ege_scores,
    )

    session.add(applicant)
    await session.flush()

    for index, direction_name in enumerate(data.priorities, start=1):
        direction = await get_direction_by_name(session, direction_name)
        if not direction:
            raise HTTPException(
                status_code=400,
                detail=f"Direction not found: {direction_name}",
            )
        session.add(
            ApplicantPriority(
                applicant_id=applicant.id,
                direction_id=direction.id,
                priority=index,
                status="pending",
            )
        )

    await session.commit()
    await session.refresh(applicant)
    return applicant


async def authenticate_user(session, login: str, password: str) -> Applicant | None:
    user = await get_by_login(session, login)
    if not user:
        user = await get_by_email(session, login)
    if not user:
        return None
    if not verify_password(password, user.password_hash):
        return None
    return user


async def ensure_staff_users(session) -> None:
    defaults = [
        {
            "login": app_config.default_admissions_login,
            "email": app_config.default_admissions_login,
            "password": app_config.default_admissions_password,
            "role": "admissions",
            "last_name": "Admin",
            "first_name": "Admissions",
            "middle_name": "Office",
        },
        {
            "login": app_config.default_analyst_login,
            "email": app_config.default_analyst_login,
            "password": app_config.default_analyst_password,
            "role": "analyst",
            "last_name": "Analyst",
            "first_name": "Admissions",
            "middle_name": "Office",
        },
    ]

    for user_data in defaults:
        existing = await get_by_login(session, user_data["login"])
        if existing:
            if existing.role != user_data["role"]:
                existing.role = user_data["role"]
            if not verify_password(user_data["password"], existing.password_hash):
                existing.password_hash = hash_password(user_data["password"])
            continue
        session.add(
            Applicant(
                last_name=user_data["last_name"],
                first_name=user_data["first_name"],
                middle_name=user_data["middle_name"],
                email=user_data["email"],
                login=user_data["login"],
                password_hash=hash_password(user_data["password"]),
                role=user_data["role"],
                total_score=0,
                sex=True,
            )
        )
    await session.commit()
