-- -----------------------------------------------------------------
-- This .db will be on harddisk and saves today minuted stats
-- -----------------------------------------------------------------
PRAGMA foreign_keys = OFF;
BEGIN TRANSACTION;

-- -----------------------------------------------------------------
-- Table structure for resumen (every minute in a day)
-- -----------------------------------------------------------------
-- id 			= autoincrement key only for DB
-- username		= 1st part in stream
-- streamname	= 2nd part in stream
-- os			= "win", "lin", "mac", "and", "other"
-- isocode		= ISO 3166-2 country code
-- time			= total seconds connected today in all
-- kilobytes	= total kilobytes transferred
-- players		= number of players connected now
-- proxies		= number of proxies used now
-- hour			= (00-23) Hour time now
-- minutes		= (00-59) Minutes time now
-- date			= today's date (yyyy-mm-dd)
-- -----------------------------------------------------------------
CREATE TABLE "resumen" (
"id"  INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
"username"  TEXT(255),
"streamname"  TEXT(255),
"os"  TEXT(7),
"isocode"  TEXT(5),
"time"  INTEGER,
"kilobytes"  INTEGER,
"players"  INTEGER,
"proxies"  INTEGER,
"hour"  INTEGER,
"minutes"  INTEGER,
"date"  TEXT(255)
);

DELETE FROM sqlite_sequence;
INSERT INTO "sqlite_sequence" VALUES('resumen',0);

COMMIT;
