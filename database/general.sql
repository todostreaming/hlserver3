-- -----------------------------------------------------------------
-- This .db will be on harddisk and saves general info
-- -----------------------------------------------------------------
PRAGMA foreign_keys = OFF;
BEGIN TRANSACTION;

-- -----------------------------------------------------------------
-- Table structure for users
-- -----------------------------------------------------------------
-- id 			= autoincrement key only for DB
-- username		= 1st part in every stream published or admin name
-- password		= panel password to enter (admins and publishers)
-- pubpass		= password for every stream publishing
-- type			= type of user (superadmin=0, admin=1, publisher=2)
-- status		= enabled(1)/disabled(0) publisher (never deleted)
-- -----------------------------------------------------------------
CREATE TABLE "users" (
"id"  INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
"username"  TEXT(255),
"password"  TEXT(255),
"pubpass"  TEXT(255),
"type"  INTEGER,
"status"  INTEGER
);

-- -----------------------------------------------------------------
-- Table structure for referer (mirrored with an internal map)
-- -----------------------------------------------------------------
-- username		= 1st part in stream
-- streamname	= 2nd part in stream
-- referrers	= (;)separated pure domains allowed (if none = all)
-- -----------------------------------------------------------------
CREATE TABLE "referer" (
"username"  TEXT(255),
"streamname"  TEXT(255),
"referrers"  TEXT(1024)
);

INSERT INTO "users" VALUES(1,'admin','admin','admin',0,1);
DELETE FROM sqlite_sequence;
INSERT INTO "sqlite_sequence" VALUES('users',1);

COMMIT;
