import re
from datetime import date
from typing import Optional

from pydantic import BaseModel, ConfigDict, Field, field_validator, model_validator


def _email_regex(v: str) -> str:
    pattern = r"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$"
    if not re.match(pattern, v):
        raise ValueError("Invalid email format")
    return v.lower()


class ApplicantRegisterSchema(BaseModel):
    """Registration schema with validation."""

    model_config = ConfigDict(populate_by_name=True)

    fullname: str
    password: str
    password_confirm: str
    birthdate: date
    phone: str
    email: str
    telegram: Optional[str] = None
    school: str
    achievements: Optional[str] = None
    priorities: list[str]
    agreement: bool = True
    ege_scores: Optional[dict[str, int]] = Field(default=None, alias="egeScores")

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
            raise ValueError("Full name is required")
        return v.strip()

    @field_validator("password")
    @classmethod
    def password_min_length(cls, v: str) -> str:
        if len(v) < 6:
            raise ValueError("Password must be at least 6 characters")
        return v

    @model_validator(mode="after")
    def passwords_match(self):
        if self.password != self.password_confirm:
            raise ValueError("Passwords do not match")
        return self

    @field_validator("priorities")
    @classmethod
    def priorities_not_empty(cls, v: list[str]) -> list[str]:
        clean = [p.strip() for p in v if p and p.strip()]
        if not clean:
            raise ValueError("Select at least one direction")
        if len(clean) > 3:
            raise ValueError("Maximum 3 directions")
        return clean[:3]

    @field_validator("agreement")
    @classmethod
    def agreement_required(cls, v: bool) -> bool:
        if not v:
            raise ValueError("Agreement is required")
        return v

    def split_fullname(self) -> tuple[str, str, str | None]:
        """Split full name into last, first, middle."""
        parts = self.fullname.strip().split(maxsplit=2)
        last_name = parts[0]
        first_name = parts[1] if len(parts) > 1 else ""
        middle_name = parts[2] if len(parts) > 2 else None
        return last_name, first_name, middle_name


ApplicantSchema = ApplicantRegisterSchema


class UserLoginSchema(BaseModel):
    """Login schema."""

    username: str
    password: str
