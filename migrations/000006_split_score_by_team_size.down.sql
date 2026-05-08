-- Reverse of 000006_split_score_by_team_size.up.sql
-- Restore the view to its post-000005 state (flat value, no team-size split).

DROP VIEW IF EXISTS user_total_score;

CREATE VIEW user_total_score AS
SELECT
    u.id,
    u.username,
    u.role,
    COALESCE(SUM(
        CASE
            WHEN (f.team_1_id = t.id AND f.result = 'team_1')
              OR (f.team_2_id = t.id AND f.result = 'team_2') THEN  f.value
            ELSE                                                    -f.value
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
GROUP BY u.id, u.username, u.role;
