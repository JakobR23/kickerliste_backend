-- =============================================================
-- Migration: fixture approval workflow
-- =============================================================

CREATE TYPE fixture_status AS ENUM ('pending', 'approved', 'rejected');

ALTER TABLE fixture
    ADD COLUMN status       fixture_status  NOT NULL DEFAULT 'pending',
    ADD COLUMN submitted_by INT             REFERENCES "user"(id);

-- All fixtures that existed before this migration are already valid.
UPDATE fixture SET status = 'approved';

CREATE INDEX IF NOT EXISTS idx_fixture_status       ON fixture (status);
CREATE INDEX IF NOT EXISTS idx_fixture_submitted_by ON fixture (submitted_by);

-- Recreate the view so only approved fixtures contribute to scores.
-- Note: u.role is present because migration 000002 has already added it.
DROP VIEW IF EXISTS user_total_score;

CREATE VIEW user_total_score AS
SELECT
    u.id,
    u.username,
    u.role,
    COALESCE(SUM(f.value), 0) AS total_score
FROM "user" u
LEFT JOIN team_member tm ON tm.user_id = u.id
LEFT JOIN team t          ON t.id = tm.team_id
LEFT JOIN fixture f       ON ((f.team_1_id = t.id AND f.result = 'team_1')
                          OR  (f.team_2_id = t.id AND f.result = 'team_2'))
                         AND f.status = 'approved'
GROUP BY u.id, u.username, u.role;
