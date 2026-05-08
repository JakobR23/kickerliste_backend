-- =============================================================
-- Migration: split fixture value by team size
-- =============================================================
-- A fixture's value is now divided equally among the members of
-- each team. A solo team (1 member) receives the full value; a
-- duo team (2 members) splits it evenly. This applies to both
-- wins (positive) and losses (negative).
--
-- team_size is computed from the current team_member rows.
-- Historical team composition is not tracked, so this reflects
-- the team's present membership.

DROP VIEW IF EXISTS user_total_score;

CREATE VIEW user_total_score AS
SELECT
    u.id,
    u.username,
    u.role,
    COALESCE(SUM(
        CASE
            WHEN (f.team_1_id = t.id AND f.result = 'team_1')
              OR (f.team_2_id = t.id AND f.result = 'team_2')
            THEN  f.value::numeric / NULLIF(team_size.member_count, 0)
            ELSE -f.value::numeric / NULLIF(team_size.member_count, 0)
        END
    ), 0)
    + COALESCE((
        SELECT SUM(sa.amount)
        FROM score_adjustment sa
        WHERE sa.user_id = u.id
    ), 0) AS total_score
FROM "user" u
LEFT JOIN team_member tm ON tm.user_id = u.id
LEFT JOIN team t          ON t.id = tm.team_id
LEFT JOIN fixture f       ON (f.team_1_id = t.id OR f.team_2_id = t.id)
                         AND f.result <> 'draw'
                         AND f.status = 'approved'
LEFT JOIN (
    SELECT team_id, COUNT(*) AS member_count
    FROM team_member
    GROUP BY team_id
) team_size ON team_size.team_id = t.id
GROUP BY u.id, u.username, u.role;
