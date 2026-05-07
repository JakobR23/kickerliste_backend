-- =============================================================
-- Migration: user roles
-- =============================================================

CREATE TYPE user_role AS ENUM ('admin', 'user');

ALTER TABLE "user"
    ADD COLUMN role user_role NOT NULL DEFAULT 'user';

-- Recreate the view to expose role alongside the existing columns.
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
LEFT JOIN fixture f       ON (f.team_1_id = t.id AND f.result = 'team_1')
                          OR (f.team_2_id = t.id AND f.result = 'team_2')
GROUP BY u.id, u.username, u.role;
