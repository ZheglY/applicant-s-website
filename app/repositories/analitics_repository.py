from sqlalchemy import func, select

from app.db.models import Applicant, ApplicantPriority, FacultyDirection


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

    status_counts = await session.execute(
        select(ApplicantPriority.status, func.count(ApplicantPriority.id))
        .group_by(ApplicantPriority.status)
    )
    status_map = {row[0] or "pending": row[1] for row in status_counts.all()}

    popular_directions = await session.execute(
        select(
            FacultyDirection.direction_name,
            func.count(ApplicantPriority.id).label("applications"),
            FacultyDirection.budget_places,
            FacultyDirection.paid_places,
        )
        .join(ApplicantPriority, ApplicantPriority.direction_id == FacultyDirection.id, isouter=True)
        .group_by(FacultyDirection.id)
        .order_by(func.count(ApplicantPriority.id).desc(), FacultyDirection.id)
    )
    direction_rows = []
    for name, applications, budget, paid in popular_directions.all():
        places = int((budget or 0) + (paid or 0))
        direction_rows.append(
            {
                "name": name,
                "applications": int(applications or 0),
                "budget_places": int(budget or 0),
                "paid_places": int(paid or 0),
                "competition": round((int(applications or 0) / int(budget or 0)), 1) if budget else 0,
                "fill_percent": min(100, round((int(applications or 0) / places) * 100)) if places else 0,
            }
        )

    return {
        "total_applicants": int(total_applicants or 0),
        "total_applications": int(total_applications or 0),
        "budget_places": int(budget_places or 0),
        "paid_places": int(paid_places or 0),
        "avg_score": float(avg_score or 0),
        "priority_counts": priority_map,
        "status_counts": status_map,
        "popular_directions": direction_rows,
    }
