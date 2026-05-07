-- =============================================================
-- Migration: match-tracking schema
-- =============================================================


-- -------------------------------------------------------------
-- TYPES
-- -------------------------------------------------------------

CREATE TYPE match_result AS ENUM ('team_1', 'team_2', 'draw');


-- -------------------------------------------------------------
-- TABLES
-- -------------------------------------------------------------

CREATE TABLE "user" (
    id          SERIAL          PRIMARY KEY,
    username    VARCHAR(255)    NOT NULL UNIQUE,
    password    TEXT            NOT NULL,
    hashsalt    TEXT            NOT NULL,
    created_at  TIMESTAMP       NOT NULL DEFAULT NOW()
);

CREATE TABLE team (
    id          SERIAL          PRIMARY KEY,
    name        VARCHAR(255),
    created_at  TIMESTAMP       NOT NULL DEFAULT NOW()
);

-- 1 or 2 members per team, enforced at the application layer
CREATE TABLE team_member (
    id          SERIAL          PRIMARY KEY,
    team_id     INT             NOT NULL REFERENCES team("id")   ON DELETE CASCADE,
    user_id     INT             NOT NULL REFERENCES "user"("id") ON DELETE CASCADE
);

CREATE TABLE fixture (
    id              SERIAL          PRIMARY KEY,
    team_1_id       INT             NOT NULL REFERENCES team("id"),
    team_2_id       INT             NOT NULL REFERENCES team("id"),
    result          match_result    NOT NULL,
    score_team_1    INT,
    score_team_2    INT,
    value           INT             NOT NULL,
    played_at       TIMESTAMP       NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_fixture_value_non_negative
        CHECK (value >= 0),

    CONSTRAINT chk_fixture_scores_consistent
        CHECK (
            (score_team_1 IS NULL AND score_team_2 IS NULL) OR
            (score_team_1 IS NOT NULL AND score_team_2 IS NOT NULL)
        ),

    CONSTRAINT chk_fixture_distinct_teams
        CHECK (team_1_id <> team_2_id)
);


-- -------------------------------------------------------------
-- INDEXES
-- -------------------------------------------------------------

CREATE INDEX        IF NOT EXISTS idx_team_member_team_id   ON team_member (team_id);
CREATE INDEX        IF NOT EXISTS idx_team_member_user_id   ON team_member (user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_team_member_unique    ON team_member (team_id, user_id);

CREATE INDEX IF NOT EXISTS idx_fixture_team_1_id    ON fixture (team_1_id);
CREATE INDEX IF NOT EXISTS idx_fixture_team_2_id    ON fixture (team_2_id);
CREATE INDEX IF NOT EXISTS idx_fixture_teams        ON fixture (team_1_id, team_2_id);
CREATE INDEX IF NOT EXISTS idx_fixture_played_at    ON fixture (played_at DESC);


-- -------------------------------------------------------------
-- VIEWS
-- -------------------------------------------------------------

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
