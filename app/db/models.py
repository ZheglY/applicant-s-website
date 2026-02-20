from sqlalchemy import Column, Integer, String, DateTime, ForeignKey, Text, Boolean
from sqlalchemy.orm import DeclarativeBase, relationship, Mapped, mapped_column
from sqlalchemy.sql import func
from datetime import datetime

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

    id: Mapped[int] = mapped_column(primary_key=True)
    last_name: Mapped[str] = mapped_column(nullable=False)
    first_name: Mapped[str] = mapped_column(nullable=False)
    middle_name: Mapped[str] = mapped_column(nullable=True)
    email: Mapped[str] = mapped_column(unique=True, nullable=False)
    phone: Mapped[str] = mapped_column(nullable=True)
    birth_date: Mapped[DateTime] = mapped_column(nullable=False)
    total_score: Mapped[int] = mapped_column(default = 0)
    role: Mapped[str] = mapped_column(nullable=False)
    sex: Mapped[bool] = mapped_column(nullable=False)
    achievements: Mapped[str] = mapped_column(nullable=True)

    status_code: Mapped[int] = mapped_column(default = 0)
    faculty_direction_id: Mapped[int] = mapped_column(ForeignKey("faculty_directions.id"), nullable=False)
    faculty_direction = relationship("FacultyDirection", back_populates="applicants")

    def __repr__(self):
        return f"<Applicant {self.last_name} {self.first_name}>"