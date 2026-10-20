-- =============================================================
-- Migration: per-user score-history read model
-- =============================================================
-- A view that lists each score-changing event for a user as a time series
-- with a running cumulative total, powering the single-player profile graph.
--
-- It MUST stay in lockstep with user_total_score (migration 000006): the
-- per-fixture delta expression below is identical, so the running_total of a
-- user's last event always equals their total_score.
--
--   * Only approved, non-draw fixtures contribute (draws are excluded).
--   * A win adds  value / current_team_size, a loss subtracts it.
--   * team_size is current membership (historical composition isn't tracked).
--   * Manual score adjustments are included verbatim (amount can be negative).
--   * Zero-delta events are omitted (they don't move the curve).
--
-- seq is the event's own table id, exposed only so callers can reproduce the
-- deterministic (at, type, seq) ordering the running total was computed with.

CREATE VIEW user_score_history AS
WITH events AS (
    -- Approved, non-draw fixtures, per team the user belongs to.
    SELECT
        u.id        AS user_id,
        f.played_at AS at,
        CASE
            WHEN (f.team_1_id = t.id AND f.result = 'team_1')
              OR (f.team_2_id = t.id AND f.result = 'team_2')
            THEN  f.value::numeric / NULLIF(team_size.member_count, 0)
            ELSE -f.value::numeric / NULLIF(team_size.member_count, 0)
        END         AS delta,
        'fixture'   AS type,
        f.id        AS fixture_id,
        NULL::text  AS reason,
        f.id        AS seq
    FROM "user" u
    JOIN team_member tm ON tm.user_id = u.id
    JOIN team t         ON t.id = tm.team_id
    JOIN fixture f      ON (f.team_1_id = t.id OR f.team_2_id = t.id)
                       AND f.result <> 'draw'
                       AND f.status = 'approved'
    JOIN (
        SELECT team_id, COUNT(*) AS member_count
        FROM team_member
        GROUP BY team_id
    ) team_size ON team_size.team_id = t.id

    UNION ALL

    -- Manual score adjustments.
    SELECT
        sa.user_id         AS user_id,
        sa.created_at      AS at,
        sa.amount::numeric AS delta,
        'adjustment'       AS type,
        NULL::int          AS fixture_id,
        sa.reason          AS reason,
        sa.id              AS seq
    FROM score_adjustment sa
)
SELECT
    user_id,
    at,
    delta,
    type,
    fixture_id,
    reason,
    seq,
    SUM(delta) OVER (
        PARTITION BY user_id
        ORDER BY at, type, seq
        ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
    ) AS running_total
FROM events
WHERE delta <> 0;
