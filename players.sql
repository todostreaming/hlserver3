PRAGMA foreign_keys = OFF;

-- ----------------------------
-- Table structure for players
-- ----------------------------
DROP TABLE IF EXISTS "main"."players";
CREATE TABLE "players" (
	"id"  INTEGER PRIMARY KEY NOT NULL,
	"rawstream"  TEXT(255),
	"ipproxy"  TEXT(255),
	"ipclient"  TEXT(255),
	"timestamp"  INTEGER,
	"time"  INTEGER,
	"kilobytes"  INTEGER,
	"total_time"  INTEGER,
	"agent" TEXT(255)
);
