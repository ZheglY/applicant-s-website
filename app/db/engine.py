from typing import Annotated

from fastapi import Depends
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine

from core.config import app_config
from db.models import Base, FacultyDirection
from db.seed import DIRECTIONS


engine = create_async_engine(app_config.db.database_url, echo=app_config.debug)
new_session = async_sessionmaker(engine, expire_on_commit=False, class_=AsyncSession)


async def get_session() -> AsyncSession:
    async with new_session() as session:
        yield session


SessionDep = Annotated[AsyncSession, Depends(get_session)]


async def setup_database() -> None:
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)


async def seed_directions() -> None:
    async with new_session() as session:
        result = await session.execute(select(FacultyDirection.id).limit(1))
        if result.scalar_one_or_none() is not None:
            return

        for data in DIRECTIONS:
            session.add(FacultyDirection(**data))
        await session.commit()


async def drop_database() -> None:
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.drop_all)
