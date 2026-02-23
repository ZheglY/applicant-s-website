import bcrypt


def _hash_password(password: str) -> str:
    """Bcrypt: пароль обрезается до 72 байт (лимит bcrypt)."""
    pwd_bytes = password.encode("utf-8")[:72]
    return bcrypt.hashpw(pwd_bytes, bcrypt.gensalt()).decode()