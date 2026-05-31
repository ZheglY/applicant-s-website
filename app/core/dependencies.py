from authx import TokenPayload
from fastapi import HTTPException, Request

from app.core.config import security
from app.db.engine import SessionDep
from app.repositories.user_repository import get_by_id


def _extract_user_id(payload: TokenPayload) -> int | None:
    user_id = payload.extra_dict.get("user_id") or getattr(payload, "subject", None) or getattr(payload, "sub", None)
    try:
        return int(user_id)
    except (TypeError, ValueError):
        return None


async def _get_current_user(request: Request, session: SessionDep, redirect_on_fail: bool):
    token = await security.get_token_from_request(
        request,
        type="access",
        optional=True,
        locations=["cookies"],
    )
    if not token:
        if redirect_on_fail:
            raise HTTPException(status_code=302, headers={"Location": "/auth/enter"})
        raise HTTPException(status_code=401, detail="Authentication required")

    payload = security.verify_token(token, verify_csrf=False)
    user_id = _extract_user_id(payload)
    if not user_id:
        if redirect_on_fail:
            raise HTTPException(status_code=302, headers={"Location": "/auth/enter"})
        raise HTTPException(status_code=401, detail="Authentication required")

    user = await get_by_id(session, user_id)
    if not user:
        if redirect_on_fail:
            raise HTTPException(status_code=302, headers={"Location": "/auth/enter"})
        raise HTTPException(status_code=401, detail="User not found")
    return user


async def get_current_user(request: Request, session: SessionDep):
    return await _get_current_user(request, session, redirect_on_fail=False)


async def get_current_user_page(request: Request, session: SessionDep):
    return await _get_current_user(request, session, redirect_on_fail=True)


async def get_optional_user(request: Request, session: SessionDep):
    token = await security.get_token_from_request(
        request,
        type="access",
        optional=True,
        locations=["cookies"],
    )
    if not token:
        return None
    try:
        payload = security.verify_token(token, verify_csrf=False)
    except Exception:
        return None

    user_id = _extract_user_id(payload)
    if not user_id:
        return None
    return await get_by_id(session, user_id)
