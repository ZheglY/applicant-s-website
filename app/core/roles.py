"""
Ролевая модель доступа.

Роли:
- student: просмотр сайта, НЕТ доступа к аналитике, НЕТ удаления
- analyst: всё как студент + доступ к аналитике
- admissions: всё как студент + доступ к удалению студентов, НЕТ аналитики
"""
from fastapi import Depends, HTTPException
from authx import TokenPayload

from core.config import security


def require_analyst(payload: TokenPayload = Depends(security.access_token_required)) -> TokenPayload:
    """Только аналитик может получить доступ."""
    role = payload.extra_dict.get("role")
    if role != "analyst":
        raise HTTPException(status_code=403, detail="Доступ разрешён только аналитикам")
    return payload


def require_admissions(payload: TokenPayload = Depends(security.access_token_required)) -> TokenPayload:
    """Только приёмная комиссия может получить доступ (например, удаление)."""
    role = payload.extra_dict.get("role")
    if role != "admissions":
        raise HTTPException(status_code=403, detail="Доступ разрешён только приёмной комиссии")
    return payload


def require_authenticated(payload: TokenPayload = Depends(security.access_token_required)) -> TokenPayload:
    """Любой авторизованный пользователь (любая роль)."""
    return payload
