"""Role-based access control."""
from fastapi import Depends, HTTPException
from authx import TokenPayload

from core.config import security


def require_analyst(payload: TokenPayload = Depends(security.access_token_required)) -> TokenPayload:
    role = payload.extra_dict.get("role")
    if role != "analyst":
        raise HTTPException(status_code=403, detail="Analyst access required")
    return payload


def require_admissions(payload: TokenPayload = Depends(security.access_token_required)) -> TokenPayload:
    role = payload.extra_dict.get("role")
    if role != "admissions":
        raise HTTPException(status_code=403, detail="Admissions access required")
    return payload


def require_staff(payload: TokenPayload = Depends(security.access_token_required)) -> TokenPayload:
    role = payload.extra_dict.get("role")
    if role not in {"admissions", "analyst"}:
        raise HTTPException(status_code=403, detail="Staff access required")
    return payload


def require_authenticated(payload: TokenPayload = Depends(security.access_token_required)) -> TokenPayload:
    return payload
