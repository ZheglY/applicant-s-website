from dataclasses import dataclass
from environs import Env
from authx import AuthX, AuthXConfig

config = AuthXConfig()
config.JWT_SECRET_KEY = "SECRET_KEY"
config.JWT_ACCESS_CSRF_COOKIE_NAME = "my_access_token"
config.JWT_TOKEN_LOCATION = ["cookies"]


security = AuthX(config=config)


@dataclass
class DatabaseConfig:
    database_url: str
    

@dataclass
class Config:
    db: DatabaseConfig
    secret_key: str
    debug: bool

def load_config(path: str) -> Config:
    env = Env()
    env.read_env(path)  # Загружаем переменные окружения из файла .env

    return Config(
        db=DatabaseConfig(database_url=env("DATABASE_URL")),
        secret_key=env("SECRET_KEY"),
        debug=env.bool("DEBUG", default=False)
    )