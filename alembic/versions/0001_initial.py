"""Initial schema

Revision ID: 0001_initial
Revises: 
Create Date: 2026-03-15 18:00:00.000000
"""

from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects import postgresql


# revision identifiers, used by Alembic.
revision = "0001_initial"
down_revision = None
branch_labels = None
depends_on = None


def upgrade() -> None:
    op.create_table(
        "faculty_directions",
        sa.Column("id", sa.Integer(), primary_key=True, nullable=False),
        sa.Column("faculty_name", sa.String(), nullable=False),
        sa.Column("direction_code", sa.String(), nullable=False),
        sa.Column("direction_name", sa.String(), nullable=False),
        sa.Column("budget_places", sa.Integer(), nullable=False, server_default="0"),
        sa.Column("paid_places", sa.Integer(), nullable=False, server_default="0"),
        sa.Column("is_full", sa.Boolean(), nullable=False, server_default=sa.text("false")),
        sa.Column("subjects", postgresql.JSONB(astext_type=sa.Text()), nullable=False, server_default=sa.text("'[]'::jsonb")),
        sa.Column("created_at", sa.DateTime(timezone=True), server_default=sa.func.now(), nullable=True),
        sa.Column("updated_at", sa.DateTime(timezone=True), nullable=True),
    )

    op.create_table(
        "applicants",
        sa.Column("id", sa.Integer(), primary_key=True, autoincrement=True, nullable=False),
        sa.Column("last_name", sa.String(), nullable=False),
        sa.Column("first_name", sa.String(), nullable=False),
        sa.Column("middle_name", sa.String(), nullable=True),
        sa.Column("email", sa.String(), nullable=False),
        sa.Column("login", sa.String(), nullable=False),
        sa.Column("password_hash", sa.String(), nullable=False),
        sa.Column("phone", sa.String(), nullable=True),
        sa.Column("telegram", sa.String(), nullable=True),
        sa.Column("birth_date", sa.Date(), nullable=True),
        sa.Column("total_score", sa.Integer(), nullable=False, server_default="0"),
        sa.Column("role", sa.String(), nullable=False, server_default="student"),
        sa.Column("sex", sa.Boolean(), nullable=False, server_default=sa.text("true")),
        sa.Column("achievements", postgresql.JSONB(astext_type=sa.Text()), nullable=True),
        sa.Column("school", sa.String(), nullable=True),
        sa.Column("region", sa.String(), nullable=True),
        sa.Column("ege_scores", postgresql.JSONB(astext_type=sa.Text()), nullable=True),
        sa.Column("created_at", sa.DateTime(timezone=True), server_default=sa.func.now(), nullable=True),
        sa.Column("updated_at", sa.DateTime(timezone=True), nullable=True),
        sa.UniqueConstraint("email"),
        sa.UniqueConstraint("login"),
    )

    op.create_table(
        "applicant_priorities",
        sa.Column("id", sa.Integer(), primary_key=True, autoincrement=True, nullable=False),
        sa.Column("applicant_id", sa.Integer(), sa.ForeignKey("applicants.id"), nullable=False),
        sa.Column("direction_id", sa.Integer(), sa.ForeignKey("faculty_directions.id"), nullable=False),
        sa.Column("priority", sa.Integer(), nullable=False),
        sa.Column("status", sa.String(), nullable=False, server_default="pending"),
        sa.Column("created_at", sa.DateTime(timezone=True), server_default=sa.func.now(), nullable=True),
        sa.Column("updated_at", sa.DateTime(timezone=True), nullable=True),
        sa.UniqueConstraint("applicant_id", "direction_id", name="uq_applicant_direction"),
    )

    op.create_table(
        "news",
        sa.Column("id", sa.Integer(), primary_key=True, autoincrement=True, nullable=False),
        sa.Column("title", sa.String(), nullable=False),
        sa.Column("subtitle", sa.String(), nullable=False),
        sa.Column("text", sa.Text(), nullable=False),
        sa.Column("image_url", sa.String(), nullable=True),
        sa.Column("author_id", sa.Integer(), sa.ForeignKey("applicants.id"), nullable=True),
        sa.Column("created_at", sa.DateTime(timezone=True), server_default=sa.func.now(), nullable=True),
        sa.Column("updated_at", sa.DateTime(timezone=True), nullable=True),
    )


def downgrade() -> None:
    op.drop_table("news")
    op.drop_table("applicant_priorities")
    op.drop_table("applicants")
    op.drop_table("faculty_directions")
