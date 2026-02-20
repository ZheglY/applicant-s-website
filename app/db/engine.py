from sqlalchemy.ext.asyncio import create_async_engine, async_sessionmaker, AsyncSession
from models import Base
from typing import Annotated
from fastapi import Depends

engine = create_async_engine('sqlite + aiosqlite:///applicants.db')

new_session = async_sessionmaker(engine, expire_on_commit=False)

async def get_session():
    async with new_session() as session:
        yield session

SessionDep = Annotated[AsyncSession, Depends(get_session())]

async def setup_database():
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)

async def drop_database():
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.drop_all)  