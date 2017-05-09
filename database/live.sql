-- -----------------------------------------------------------------
-- This .db will be on RAMdisk and backed up to disk periodically
-- -----------------------------------------------------------------
PRAGMA foreign_keys = OFF;
BEGIN TRANSACTION;

-- -----------------------------------------------------------------
-- Table structure for players
-- -----------------------------------------------------------------
-- id (64 bits) = variant playlist from wid%d
-- username		= 1st part in stream
-- streamname	= 2nd part in stream
-- os			= "win", "lin", "mac", "and", "other"
-- ipproxy		= ip from the near proxy CDN
-- ipclient		= remote ip
-- isocode		= ISO 3166-2 country code
-- country		= complete country name
-- city			= complete name of city
-- timestamp	= UNIX 64 bits timestamp of latest update
-- time			= seconds connected since the latest disconnection
-- kilobytes	= total kilobytes transferred
-- total_time	= total seconds connected today in all
-- -----------------------------------------------------------------
CREATE TABLE "players" (
"id"  INTEGER PRIMARY KEY NOT NULL,
"username"  TEXT(255),
"streamname"  TEXT(255),
"os"  TEXT(7),
"ipproxy"  TEXT(255),
"ipclient"  TEXT(255),
"isocode"  TEXT(4),
"country"  TEXT(255),
"city"  TEXT(255),
"timestamp"  INTEGER,
"time"  INTEGER,
"kilobytes"  INTEGER,
"total_time"  INTEGER
);

-- -----------------------------------------------------------------
-- Table structure for encoders
-- -----------------------------------------------------------------
-- id 			= autoincrement key only for DB
-- username		= 1st part in stream
-- streamname	= 2nd part in stream
-- time			= seconds connected since the latest disconnection
-- bitrate		= kbps of publishing
-- ip			= ip of publisher
-- info			= codec and general video info
-- isocode		= ISO 3166-2 country code
-- country		= complete country name
-- city			= complete name of city
-- timestamp	= UNIX 64 bits timestamp of latest update
-- -----------------------------------------------------------------
CREATE TABLE "encoders" (
"id"  INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
"username"  TEXT(255),
"streamname"  TEXT(255),
"time"  INTEGER,
"bitrate"  INTEGER,
"ip"  TEXT(255),
"info"  TEXT(255),
"isocode"  TEXT(4),
"country"  TEXT(255),
"city"  TEXT(255),
"timestamp"  INTEGER
);

DELETE FROM sqlite_sequence;
INSERT INTO "sqlite_sequence" VALUES('encoders',0);

COMMIT;
