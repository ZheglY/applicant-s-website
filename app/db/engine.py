from typing import Annotated

from fastapi import Depends
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine

from db.models import Base
from db.models import FacultyDirection
from db.seed import DIRECTIONS


# Создание движка и сессии (echo=False - отключает логирование SQL запросов)
# Драйвер aiosqlite — асинхронный | файл базы: applicants.db | 
engine = create_async_engine("sqlite+aiosqlite:///applicants.db", echo=False)

# создание сессии / expire_on_commit=False — объекты не инвалидируются после commit / class_=AsyncSession — используем асинхронную сессию
new_session = async_sessionmaker(engine, expire_on_commit=False, class_=AsyncSession)


async def get_session() -> AsyncSession:
    """
    Dependency-функция. FastAPI будет: 
    1. Создавать сессию
    2. Передавать её в качестве зависимости в эндпоинты
    3. Закрывать сессию после запроса автоматически
    """
    async with new_session() as session:
        yield session


# Теперь можно писать в роутере: async def route(session: SessionDep):
SessionDep = Annotated[AsyncSession, Depends(get_session)]

async def setup_database():
    """
    1. Открывает соединение
    2. Вызывает create_all()
    3. Создаёт таблицы если их нет
    """
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)


async def seed_directions():
    """
    Функция для заполнения таблицы направлениями при запуске
    """

    async with new_session() as session: # Проверка: есть ли данные
        result = await session.execute(select(FacultyDirection.id).limit(1))
        if result.scalar_one_or_none() is not None:
            return  # Если в таблице уже есть хотя бы одна запись → ничего не делаем.

        # Заполнение таблицы направлениями факультетов если данных нет 
        for fac_name, code, dir_name in DIRECTIONS:
            session.add(FacultyDirection(
                faculty_name=fac_name, direction_code=code, direction_name=dir_name
            ))
        await session.commit()


async def drop_database():
    """
    Удаляет ВСЕ таблицы.
    """
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.drop_all)  
