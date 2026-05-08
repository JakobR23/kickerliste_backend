-- =============================================================
-- Migration: score adjustments + loss deductions
-- =============================================================

CREATE TABLE score_adjustment (
    id          SERIAL      PRIMARY KEY,
    user_id     INT         NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    amount      INT         NOT NULL,   -- positive = add points, negative = deduct
    reason      TEXT        NOT NULL,
    created_by  INT         NOT NULL REFERENCES "user"(id),
    created_at  TIMESTAMP   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_score_adjustment_user_id ON score_adjustment (user_id);

-- Recreate the view to:
--   1. Deduct fixture.value for losses as well as adding it for wins.
--   2. Add manual score adjustments from the score_adjustment table.
-- Draws (result = 'draw') are excluded from the join and contribute 0.
-- Only approved fixtures count (f.status added by migration 000003).
-- Note: u.role is present because migration 000002 has already added it.
DROP VIEW IF EXISTS user_total_score;

CREATE VIEW user_total_score AS
SELECT
    u.id,
    u.username,
    u.role,
    COALESCE(SUM(
        CASE
            WHEN f.team_1_id = t.id AND f.result = 'team_1' THEN  f.value
            WHEN f.team_2_id = t.id AND f.result = 'team_2' THEN  f.value
            WHEN f.team_1_id = t.id AND f.result = 'team_2' THEN -f.value
            WHEN f.team_2_id = t.id AND f.result = 'team_1' THEN -f.value
            ELSE 0
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
