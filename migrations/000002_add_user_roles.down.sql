-- Reverse of 000002_add_user_roles.up.sql

-- Restore the view without the role column.
DROP VIEW IF EXISTS user_total_score;

CREATE VIEW user_total_score AS
SELECT
    u.id,
    u.username,
    COALESCE(SUM(f.value), 0) AS total_score
FROM "user" u
LEFT JOIN team_member tm ON tm.user_id = u.id
LEFT JOIN team t          ON t.id = tm.team_id
LEFT JOIN fixture f       ON (f.team_1_id = t.id AND f.result = 'team_1')
                          OR (f.team_2_id = t.id AND f.result = 'team_2')
GROUP BY u.id, u.username;

ALTER TABLE "user" DROP COLUMN role;

DROP TYPE IF EXISTS user_role;
