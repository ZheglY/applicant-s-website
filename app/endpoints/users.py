from datetime import date

from fastapi import APIRouter, Depends, HTTPException, Request
from fastapi.responses import HTMLResponse
from sqlalchemy import select

from app.core.dependencies import get_current_user, get_current_user_page, get_optional_user
from app.core.templates import templates
from app.db.engine import SessionDep
from app.db.models import Applicant, ApplicantPriority, News
from app.repositories.analitics_repository import get_summary
from app.schemas.user import ApplicantStatusUpdateSchema, NewsCreateSchema
from app.repositories.user_repository import (
    get_applicant_with_priorities,
    get_by_id,
    list_directions,
    set_priority_status,
)


router = APIRouter(
    prefix="/users",
    tags=["users"],
)

STATUS_LABELS = {
    "accepted": "Зачислен",
    "pending": "На рассмотрении",
    "rejected": "Отклонён",
}

STATUS_CLASSES = {
    "accepted": "status_accepted",
    "pending": "status_pending",
    "rejected": "status_rejected",
}

PROFILE_STATUS_CLASSES = {
    "accepted": "accepted",
    "pending": "pending",
    "rejected": "rejected",
}

SUBJECT_ALIASES = {
    "Math": ["Math", "Mathematics", "Mathematics (profile)", "Математика", "Математика (проф.)"],
    "Russian": ["Russian", "Русский язык"],
    "Informatics": ["Informatics", "Информатика"],
    "Physics": ["Physics", "Физика"],
    "Social Studies": ["Social Studies", "Обществознание"],
    "Web Design": ["Web Design", "Веб-дизайн", "Дизайн"],
    "Networks": ["Networks", "Сети"],
    "Linux": ["Linux"],
    "Drawing": ["Drawing", "Рисунок"],
    "Composition": ["Composition", "Композиция"],
}


def _calc_age(birth_date: date | None) -> int | None:
    if not birth_date:
        return None
    today = date.today()
    return today.year - birth_date.year - ((today.month, today.day) < (birth_date.month, birth_date.day))


def _score_for_subject(ege_scores: dict | None, subject: str) -> int | None:
    if not ege_scores:
        return None
    aliases = SUBJECT_ALIASES.get(subject, [subject])
    for alias in aliases:
        for key, value in ege_scores.items():
            if key.lower() == alias.lower():
                return value
    return None


def _competitive_score(applicant: Applicant, subjects: list[str] | None) -> int:
    score = 0
    for subject in subjects or []:
        score += _score_for_subject(applicant.ege_scores, subject) or 0
    return score


@router.get("/news", response_class=HTMLResponse)
async def news(request: Request, current_user=Depends(get_current_user_page)):
    can_manage_news = current_user.role == "admissions"
    return templates.TemplateResponse(
        "main.html",
        {
            "request": request,
            "current_user": current_user,
            "can_manage_news": can_manage_news,
        },
    )


@router.get("/news/data")
async def news_data(session: SessionDep, current_user=Depends(get_current_user)):
    result = await session.execute(select(News).order_by(News.created_at.desc()))
    items = []
    for news in result.scalars().all():
        items.append(
            {
                "id": news.id,
                "title": news.title,
                "subtitle": news.subtitle,
                "text": news.text,
                "image": news.image_url,
                "date": news.created_at.strftime("%d.%m.%Y") if news.created_at else "",
            }
        )
    return {"items": items}


@router.post("/news")
async def create_news(
    payload: NewsCreateSchema,
    session: SessionDep,
    current_user=Depends(get_current_user_page),
):
    if current_user.role != "admissions":
        raise HTTPException(status_code=403, detail="Admissions only")

    news = News(
        title=payload.title,
        subtitle=payload.subtitle,
        text=payload.text,
        image_url=payload.image,
        author_id=current_user.id,
    )
    session.add(news)
    await session.commit()
    await session.refresh(news)

    return {
        "id": news.id,
        "title": news.title,
        "subtitle": news.subtitle,
        "text": news.text,
        "image": news.image_url,
        "date": news.created_at.strftime("%d.%m.%Y") if news.created_at else "",
    }


@router.delete("/news/{news_id}")
async def delete_news(
    news_id: int,
    session: SessionDep,
    current_user=Depends(get_current_user_page),
):
    if current_user.role != "admissions":
        raise HTTPException(status_code=403, detail="Admissions only")
    result = await session.execute(select(News).where(News.id == news_id))
    news = result.scalar_one_or_none()
    if not news:
        raise HTTPException(status_code=404, detail="News not found")
    await session.delete(news)
    await session.commit()
    return {"message": "Deleted"}


@router.get("/list", response_class=HTMLResponse)
async def list_page(request: Request, session: SessionDep, current_user=Depends(get_optional_user)):
    if not current_user:
        current_user = type("GuestUser", (), {"id": None, "role": "guest", "full_name": "Гость"})()
    directions = await list_directions(session)
    direction_views = []

    for direction in directions:
        result = await session.execute(
            select(ApplicantPriority, Applicant)
            .join(Applicant, Applicant.id == ApplicantPriority.applicant_id)
            .where(ApplicantPriority.direction_id == direction.id)
        )
        applicants = []
        for priority, applicant in result.all():
            scores = [
                _score_for_subject(applicant.ege_scores, subject)
                for subject in direction.subjects
            ]
            competitive_score = _competitive_score(applicant, direction.subjects)
            status = priority.status or "pending"
            applicants.append(
                {
                    "id": applicant.id,
                    "full_name": applicant.full_name,
                    "age": _calc_age(applicant.birth_date),
                    "scores": scores,
                    "total_score": competitive_score,
                    "status": status,
                    "status_label": STATUS_LABELS.get(status, "Pending"),
                    "status_class": STATUS_CLASSES.get(status, "status_pending"),
                }
            )
        applicants.sort(key=lambda item: (-item["total_score"], item["id"]))

        direction_views.append(
            {
                "id": direction.id,
                "name": direction.direction_name,
                "budget_places": direction.budget_places,
                "paid_places": direction.paid_places,
                "subjects": direction.subjects,
                "applicants": applicants,
            }
        )

    return templates.TemplateResponse(
        "list.html",
        {
            "request": request,
            "current_user": current_user,
            "directions": direction_views,
        },
    )


@router.get("/stats", response_class=HTMLResponse)
async def stats_page(request: Request, session: SessionDep, current_user=Depends(get_current_user_page)):
    if current_user.role not in {"admissions", "analyst"}:
        raise HTTPException(status_code=403, detail="Staff access required")

    summary = await get_summary(session)
    total_applications = summary["total_applications"]
    total_places = summary["budget_places"] + summary["paid_places"]
    avg_score = summary["avg_score"]
    competition = (total_applications / summary["budget_places"]) if summary["budget_places"] else 0
    plan_percent = min(100, round((total_applications / total_places) * 100)) if total_places else 0

    priority_counts = summary["priority_counts"]

    return templates.TemplateResponse(
        "stats.html",
        {
            "request": request,
            "current_user": current_user,
            "total_applicants": summary["total_applicants"],
            "total_applications": total_applications,
            "budget_places": summary["budget_places"],
            "paid_places": summary["paid_places"],
            "avg_score": round(avg_score, 1),
            "competition": round(competition, 1),
            "plan_total": total_places,
            "plan_percent": plan_percent,
            "priority1": priority_counts.get(1, 0),
            "priority2": priority_counts.get(2, 0),
            "priority3": priority_counts.get(3, 0),
            "accepted_count": summary["status_counts"].get("accepted", 0),
            "pending_count": summary["status_counts"].get("pending", 0),
            "rejected_count": summary["status_counts"].get("rejected", 0),
            "popular_directions": summary["popular_directions"],
        },
    )


@router.get("/applicants/list")
async def show_applicants_list(session: SessionDep, current_user=Depends(get_current_user)):
    if current_user.role not in {"admissions", "analyst"}:
        raise HTTPException(status_code=403, detail="Access denied")
    result = await session.execute(
        select(Applicant).where(Applicant.role == "student").order_by(Applicant.id)
    )
    applicants = []
    for applicant in result.scalars().all():
        applicants.append(
            {
                "id": applicant.id,
                "full_name": applicant.full_name,
                "total_score": applicant.total_score,
            }
        )
    return {"items": applicants}


@router.get("/applicants/{student_id}", response_class=HTMLResponse)
async def get_applicant_account(
    request: Request,
    student_id: int,
    session: SessionDep,
    current_user=Depends(get_current_user_page),
):
    if current_user.role == "student" and current_user.id != student_id:
        raise HTTPException(status_code=403, detail="Access denied")

    applicant = await get_applicant_with_priorities(session, student_id)
    if not applicant:
        raise HTTPException(status_code=404, detail="Applicant not found")

    priorities = sorted(applicant.priorities, key=lambda p: p.priority)
    directions_view = []

    for priority in priorities:
        direction = priority.direction
        result = await session.execute(
            select(ApplicantPriority, Applicant)
            .join(Applicant, Applicant.id == ApplicantPriority.applicant_id)
            .where(ApplicantPriority.direction_id == direction.id)
        )
        ranked_applicants = [
            {
                "id": item.id,
                "score": _competitive_score(item, direction.subjects),
            }
            for _, item in result.all()
        ]
        ranked_applicants.sort(key=lambda item: (-item["score"], item["id"]))
        position = next(
            (index for index, item in enumerate(ranked_applicants, start=1) if item["id"] == applicant.id),
            1,
        )
        position_text = f"{position} из {len(ranked_applicants)}"
        place_type = "Budget" if position <= direction.budget_places else "Paid"
        competitive_score = _competitive_score(applicant, direction.subjects)

        status = priority.status or "pending"
        directions_view.append(
            {
                "priority": priority.priority,
                "name": direction.direction_name,
                "status": status,
                "status_label": STATUS_LABELS.get(status, "Pending"),
                "status_class": PROFILE_STATUS_CLASSES.get(status, "pending"),
                "score": competitive_score,
                "position": position_text,
                "place_type": "Бюджет" if place_type == "Budget" else "Платное",
                "reason": "Высокий конкурс" if status == "rejected" else "",
                "direction_id": direction.id,
            }
        )

    achievements = applicant.achievements or []
    bonus_points = sum(item.get("points", 0) for item in achievements)

    ege_scores_list = []
    if applicant.ege_scores:
        for subject, score in applicant.ege_scores.items():
            ege_scores_list.append({"subject": subject, "score": score})

    return templates.TemplateResponse(
        "lk.html",
        {
            "request": request,
            "current_user": current_user,
            "applicant": applicant,
            "age": _calc_age(applicant.birth_date),
            "ege_scores": ege_scores_list,
            "total_score": applicant.total_score,
            "achievements": achievements,
            "bonus_points": bonus_points,
            "directions": directions_view,
            "contacts": {
                "email": applicant.email,
                "phone": applicant.phone,
                "telegram": applicant.telegram,
            },
            "can_edit_status": current_user.role == "admissions",
            "primary_direction_id": directions_view[0]["direction_id"] if directions_view else None,
        },
    )


@router.patch("/applicants/{student_id}/status")
async def update_status(
    student_id: int,
    payload: ApplicantStatusUpdateSchema,
    session: SessionDep,
    current_user=Depends(get_current_user_page),
):
    if current_user.role != "admissions":
        raise HTTPException(status_code=403, detail="Admissions only")

    applicant = await get_by_id(session, student_id)
    if not applicant:
        raise HTTPException(status_code=404, detail="Applicant not found")

    updated = await set_priority_status(session, student_id, payload.direction_id, payload.status)
    if not updated:
        raise HTTPException(status_code=404, detail="Priority not found")
    return {"message": "Status updated"}
