CREATE TABLE IF NOT EXISTS faculty_directions (
    id BIGSERIAL PRIMARY KEY,
    faculty_name TEXT NOT NULL,
    direction_code TEXT NOT NULL,
    direction_name TEXT NOT NULL,
    budget_places INTEGER NOT NULL DEFAULT 0 CHECK (budget_places >= 0),
    paid_places INTEGER NOT NULL DEFAULT 0 CHECK (paid_places >= 0),
    is_full BOOLEAN NOT NULL DEFAULT FALSE,
    subjects JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_faculty_directions_code
    ON faculty_directions (direction_code);
CREATE UNIQUE INDEX IF NOT EXISTS ux_faculty_directions_name
    ON faculty_directions (direction_name);

CREATE TABLE IF NOT EXISTS applicants (
    id BIGSERIAL PRIMARY KEY,
    last_name TEXT NOT NULL,
    first_name TEXT NOT NULL,
    middle_name TEXT,
    email TEXT NOT NULL UNIQUE,
    login TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    phone TEXT,
    telegram TEXT,
    birth_date DATE,
    total_score INTEGER NOT NULL DEFAULT 0,
    role TEXT NOT NULL DEFAULT 'student',
    sex BOOLEAN NOT NULL DEFAULT TRUE,
    achievements JSONB,
    school TEXT,
    region TEXT,
    ege_scores JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS ix_applicants_role ON applicants (role);

CREATE TABLE IF NOT EXISTS applicant_priorities (
    id BIGSERIAL PRIMARY KEY,
    applicant_id BIGINT NOT NULL REFERENCES applicants(id) ON DELETE CASCADE,
    direction_id BIGINT NOT NULL REFERENCES faculty_directions(id) ON DELETE RESTRICT,
    priority INTEGER NOT NULL CHECK (priority BETWEEN 1 AND 3),
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    CONSTRAINT uq_applicant_direction UNIQUE (applicant_id, direction_id)
);

CREATE INDEX IF NOT EXISTS ix_applicant_priorities_direction
    ON applicant_priorities (direction_id);
CREATE INDEX IF NOT EXISTS ix_applicant_priorities_applicant
    ON applicant_priorities (applicant_id);

CREATE TABLE IF NOT EXISTS news (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    subtitle TEXT NOT NULL,
    text TEXT NOT NULL,
    image_url TEXT,
    author_id BIGINT REFERENCES applicants(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS ix_news_created_at ON news (created_at DESC);
