from sqlalchemy import func, select

from db.models import Applicant, ApplicantPriority, FacultyDirection


async def get_summary(session) -> dict:
    total_applicants = await session.scalar(
        select(func.count(Applicant.id)).where(Applicant.role == "student")
    )
    total_applications = await session.scalar(select(func.count(ApplicantPriority.id)))
    budget_places = await session.scalar(select(func.coalesce(func.sum(FacultyDirection.budget_places), 0)))
    paid_places = await session.scalar(select(func.coalesce(func.sum(FacultyDirection.paid_places), 0)))
    avg_score = await session.scalar(select(func.coalesce(func.avg(Applicant.total_score), 0)).where(Applicant.role == "student"))

    priority_counts = await session.execute(
        select(ApplicantPriority.priority, func.count(ApplicantPriority.id))
        .group_by(ApplicantPriority.priority)
    )
    priority_map = {row[0]: row[1] for row in priority_counts.all()}

    return {
        "total_applicants": int(total_applicants or 0),
        "total_applications": int(total_applications or 0),
        "budget_places": int(budget_places or 0),
        "paid_places": int(paid_places or 0),
        "avg_score": float(avg_score or 0),
        "priority_counts": priority_map,
    }
