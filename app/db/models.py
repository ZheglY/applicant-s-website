from datetime import date

from sqlalchemy import Boolean, Date, DateTime, ForeignKey, UniqueConstraint
from sqlalchemy.dialects.postgresql import JSONB
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column, relationship
from sqlalchemy.sql import func


class Base(DeclarativeBase):
    created_at = mapped_column(DateTime(timezone=True), server_default=func.now())
    updated_at = mapped_column(DateTime(timezone=True), onupdate=func.now())


class FacultyDirection(Base):
    __tablename__ = "faculty_directions"

    id: Mapped[int] = mapped_column(primary_key=True)
    faculty_name: Mapped[str] = mapped_column(nullable=False)
    direction_code: Mapped[str] = mapped_column(nullable=False)
    direction_name: Mapped[str] = mapped_column(nullable=False)
    budget_places: Mapped[int] = mapped_column(default=0)
    paid_places: Mapped[int] = mapped_column(default=0)
    is_full: Mapped[bool] = mapped_column(Boolean, default=False)
    subjects: Mapped[list[str]] = mapped_column(JSONB, default=list)

    priorities = relationship("ApplicantPriority", back_populates="direction")


class Applicant(Base):
    __tablename__ = "applicants"

    id: Mapped[int] = mapped_column(primary_key=True, autoincrement=True)
    last_name: Mapped[str] = mapped_column(nullable=False)
    first_name: Mapped[str] = mapped_column(nullable=False)
    middle_name: Mapped[str | None] = mapped_column(nullable=True)
    email: Mapped[str] = mapped_column(unique=True, nullable=False)
    login: Mapped[str] = mapped_column(unique=True, nullable=False)
    password_hash: Mapped[str] = mapped_column(nullable=False)
    phone: Mapped[str | None] = mapped_column(nullable=True)
    telegram: Mapped[str | None] = mapped_column(nullable=True)
    birth_date: Mapped[date | None] = mapped_column(Date, nullable=True)
    total_score: Mapped[int] = mapped_column(default=0)
    role: Mapped[str] = mapped_column(nullable=False, default="student")
    sex: Mapped[bool] = mapped_column(Boolean, nullable=False, default=True)
    achievements: Mapped[list[dict] | None] = mapped_column(JSONB, nullable=True)
    school: Mapped[str | None] = mapped_column(nullable=True)
    region: Mapped[str | None] = mapped_column(nullable=True)
    ege_scores: Mapped[dict[str, int] | None] = mapped_column(JSONB, nullable=True)

    priorities = relationship("ApplicantPriority", back_populates="applicant", cascade="all, delete-orphan")
    news = relationship("News", back_populates="author")

    @property
    def full_name(self) -> str:
        parts = [self.last_name, self.first_name, self.middle_name]
        return " ".join([p for p in parts if p])


class ApplicantPriority(Base):
    __tablename__ = "applicant_priorities"
    __table_args__ = (
        UniqueConstraint("applicant_id", "direction_id", name="uq_applicant_direction"),
    )

    id: Mapped[int] = mapped_column(primary_key=True, autoincrement=True)
    applicant_id: Mapped[int] = mapped_column(ForeignKey("applicants.id"), nullable=False)
    direction_id: Mapped[int] = mapped_column(ForeignKey("faculty_directions.id"), nullable=False)
    priority: Mapped[int] = mapped_column(nullable=False)
    status: Mapped[str] = mapped_column(default="pending")

    applicant = relationship("Applicant", back_populates="priorities")
    direction = relationship("FacultyDirection", back_populates="priorities")


class News(Base):
    __tablename__ = "news"

    id: Mapped[int] = mapped_column(primary_key=True, autoincrement=True)
    title: Mapped[str] = mapped_column(nullable=False)
    subtitle: Mapped[str] = mapped_column(nullable=False)
    text: Mapped[str] = mapped_column(nullable=False)
    image_url: Mapped[str | None] = mapped_column(nullable=True)
    author_id: Mapped[int | None] = mapped_column(ForeignKey("applicants.id"), nullable=True)

    author = relationship("Applicant", back_populates="news")
