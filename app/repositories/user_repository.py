from sqlalchemy import select, update
from sqlalchemy.orm import selectinload

from app.db.models import Applicant, ApplicantPriority, FacultyDirection


async def get_by_id(session, user_id: int) -> Applicant | None:
    result = await session.execute(
        select(Applicant).where(Applicant.id == user_id)
    )
    return result.scalar_one_or_none()


async def get_by_email(session, email: str) -> Applicant | None:
    result = await session.execute(
        select(Applicant).where(Applicant.email == email)
    )
    return result.scalar_one_or_none()


async def get_by_login(session, login: str) -> Applicant | None:
    result = await session.execute(
        select(Applicant).where(Applicant.login == login)
    )
    return result.scalar_one_or_none()


async def create_applicant(session, applicant: Applicant) -> Applicant:
    session.add(applicant)
    await session.commit()
    await session.refresh(applicant)
    return applicant


async def get_direction_by_name(session, direction_name: str) -> FacultyDirection | None:
    result = await session.execute(
        select(FacultyDirection).where(
            FacultyDirection.direction_name == direction_name
        )
    )
    return result.scalar_one_or_none()


async def list_directions(session) -> list[FacultyDirection]:
    result = await session.execute(select(FacultyDirection).order_by(FacultyDirection.id))
    return list(result.scalars().all())


async def get_applicant_with_priorities(session, user_id: int) -> Applicant | None:
    result = await session.execute(
        select(Applicant)
        .options(selectinload(Applicant.priorities).selectinload(ApplicantPriority.direction))
        .where(Applicant.id == user_id)
    )
    return result.scalar_one_or_none()


async def set_priority_status(
    session,
    applicant_id: int,
    direction_id: int,
    status: str,
) -> bool:
    result = await session.execute(
        update(ApplicantPriority)
        .where(
            ApplicantPriority.applicant_id == applicant_id,
            ApplicantPriority.direction_id == direction_id,
        )
        .values(status=status)
    )
    await session.commit()
    return (result.rowcount or 0) > 0
