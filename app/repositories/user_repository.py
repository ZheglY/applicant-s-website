from sqlalchemy import select
from db.models import Applicant, FacultyDirection


"""CRUD функции для работы с базой данных"""

async def get_by_id(session, id: str):
    """Получение пользователя по ID"""
    result = await session.execute(
        select(Applicant).where(Applicant.id == id)
    )
    return result.scalar_one_or_none()


async def get_by_email(session, email: str):
    """Получение пользователя по ID"""
    result = await session.execute(
        select(Applicant).where(Applicant.email == email)
    )
    return result.scalar_one_or_none()


async def create_applicant(session, applicant: Applicant):
    """Добавление нового абитуриента"""
    session.add(applicant)
    await session.commint()
    await session.refresh(applicant)
    return applicant


async def get_direction_by_name(session, direction_name):
    """Получение информации о факультете по его названию"""
    result = await session.execute(
        select(FacultyDirection).where(
            FacultyDirection.direction_name == direction_name
        )
    )
    return result.scalar_one_or_none()