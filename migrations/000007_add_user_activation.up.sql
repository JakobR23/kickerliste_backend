ALTER TABLE "user" ADD COLUMN active BOOLEAN NOT NULL DEFAULT FALSE;

-- Grandfather all existing accounts as active so no current user is locked out.
UPDATE "user" SET active = TRUE;
