from datetime import datetime

from sqlalchemy import Column, DateTime, ForeignKey
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column, relationship
from sqlalchemy.sql import func

class Base(DeclarativeBase):
    created_at = Column(DateTime(timezone=True), server_default = func.now())
    updated_at = Column(DateTime(timezone=True), onupdate = func.now())


class FacultyDirection(Base):
    """
    SQLAlchemy model for facultys directions
    """

    __tablename__ = "faculty_directions"

    id: Mapped[int] = mapped_column(primary_key=True)
    faculty_name: Mapped[str] = mapped_column(nullable=False)
    direction_code: Mapped[str] = mapped_column(nullable=False)
    direction_name: Mapped[str] = mapped_column(nullable=False)
    budget_places: Mapped[int] = mapped_column(default=0)
    paid_places: Mapped[int] = mapped_column(default = 0)
    is_full: Mapped[bool] = mapped_column(default=False)

    applicants = relationship("Applicant", back_populates="faculty_direction")

class Applicant(Base):
    """
    SQLAlchemy model for applicants
    """
    __tablename__ = "applicants"

    id: Mapped[int] = mapped_column(primary_key=True, autoincrement=True)
    last_name: Mapped[str] = mapped_column(nullable=False)
    first_name: Mapped[str] = mapped_column(nullable=False)
    middle_name: Mapped[str] = mapped_column(nullable=True)
    email: Mapped[str] = mapped_column(unique=True, nullable=False)
    login: Mapped[str] = mapped_column(unique=True, nullable=False)  # для входа, = email
    password_hash: Mapped[str] = mapped_column(nullable=False)
    phone: Mapped[str] = mapped_column(nullable=True)
    telegram: Mapped[str] = mapped_column(nullable=True)
    birth_date: Mapped[datetime] = mapped_column(DateTime, nullable=False)
    total_score: Mapped[int] = mapped_column(default=0)
    role: Mapped[str] = mapped_column(nullable=False, default="student")
    sex: Mapped[bool] = mapped_column(nullable=False, default=True)  # True=М, False=Ж
    achievements: Mapped[str] = mapped_column(nullable=True)
    school: Mapped[str] = mapped_column(nullable=False)

    status_code: Mapped[int] = mapped_column(default=0)
    faculty_direction_id: Mapped[int] = mapped_column(ForeignKey("faculty_directions.id"), nullable=False)
    faculty_direction = relationship("FacultyDirection", back_populates="applicants")

    def __repr__(self):
        return f"<Applicant {self.last_name} {self.first_name}>"