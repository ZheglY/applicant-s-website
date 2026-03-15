from dataclasses import dataclass
from pathlib import Path
from typing import Optional

from authx import AuthX, AuthXConfig
from environs import Env


BASE_DIR = Path(__file__).resolve().parents[2]
DEFAULT_ENV_PATH = BASE_DIR / ".env"


@dataclass
class DatabaseConfig:
    database_url: str


@dataclass
class Config:
    db: DatabaseConfig
    secret_key: str
    debug: bool
    default_admissions_login: str
    default_admissions_password: str
    default_analyst_login: str
    default_analyst_password: str


def load_config(path: Optional[str | Path] = None) -> Config:
    env = Env()
    env_path = Path(path) if path else DEFAULT_ENV_PATH
    if env_path.exists():
        env.read_env(str(env_path))
    else:
        env.read_env()

    return Config(
        db=DatabaseConfig(
            database_url=env.str("DATABASE_URL")
        ),
        secret_key=env.str("SECRET_KEY", default="dev-secret"),
        debug=env.bool("DEBUG", default=False),
        default_admissions_login=env.str("DEFAULT_ADMISSIONS_LOGIN", default="admin@unik.edu"),
        default_admissions_password=env.str("DEFAULT_ADMISSIONS_PASSWORD", default="admin"),
        default_analyst_login=env.str("DEFAULT_ANALYST_LOGIN", default="prepod@unik.edu"),
        default_analyst_password=env.str("DEFAULT_ANALYST_PASSWORD", default="123456"),
    )


app_config = load_config()

auth_config = AuthXConfig()
auth_config.JWT_SECRET_KEY = app_config.secret_key
auth_config.JWT_ACCESS_COOKIE_NAME = "access_token"
auth_config.JWT_ACCESS_CSRF_COOKIE_NAME = "access_token_csrf"
auth_config.JWT_TOKEN_LOCATION = ["cookies"]
auth_config.JWT_COOKIE_CSRF_PROTECT = False

security = AuthX(config=auth_config)
