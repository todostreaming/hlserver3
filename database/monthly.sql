-- -----------------------------------------------------------------
-- This .db will be on harddisk and saves the month daylied stats
-- -----------------------------------------------------------------
PRAGMA foreign_keys = OFF;
BEGIN TRANSACTION;

-- -----------------------------------------------------------------
-- Table structure for resumen (every day in a month)
-- -----------------------------------------------------------------
-- id 			= autoincrement key only for DB
-- username		= 1st part in stream
-- streamname	= 2nd part in stream
-- players		= total players connected in the day
-- minutes		= total minutes played today
-- peak			= maximum number of players today
-- peaktime		= hh:mm when peak reached
-- gigabytes	= total GBs consumed in the day
-- date			= today's date (yyyy-mm-dd)
-- -----------------------------------------------------------------
CREATE TABLE "resumen" (
"id"  INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
"username"  TEXT(255),
"streamname"  TEXT(255),
"players"  INTEGER,
"minutes"  INTEGER,
"peak"  INTEGER,
"peaktime"  TEXT(10),
"gigabytes"  INTEGER,
"date"  TEXT(255)
);

DELETE FROM sqlite_sequence;
INSERT INTO "sqlite_sequence" VALUES('resumen',0);

COMMIT;
