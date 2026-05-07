-- Reverse of 000001_init_schema.up.sql
-- Drop in reverse dependency order.

DROP VIEW  IF EXISTS user_total_score;
DROP TABLE IF EXISTS fixture;
DROP TABLE IF EXISTS team_member;
DROP TABLE IF EXISTS team;
DROP TABLE IF EXISTS "user";
DROP TYPE  IF EXISTS match_result;
