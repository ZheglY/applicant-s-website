DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'ck_applicants_role'
    ) THEN
        ALTER TABLE applicants
            ADD CONSTRAINT ck_applicants_role
            CHECK (role IN ('student', 'admissions', 'analyst'));
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'ck_applicants_total_score'
    ) THEN
        ALTER TABLE applicants
            ADD CONSTRAINT ck_applicants_total_score CHECK (total_score >= 0);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'ck_applicant_priorities_status'
    ) THEN
        ALTER TABLE applicant_priorities
            ADD CONSTRAINT ck_applicant_priorities_status
            CHECK (status IN ('accepted', 'pending', 'rejected'));
    END IF;
END
$$;
