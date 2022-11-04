CREATE TABLE IF NOT EXISTS whophone (
	id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	code INTEGER NOT NULL,
	"begin" INTEGER NOT NULL,
	"end" INTEGER NOT NULL,
	capacity INTEGER NOT NULL
, operator TEXT, region TEXT, inn TEXT);

CREATE INDEX whophone_code_IDX ON whophone (code);

CREATE UNIQUE INDEX whophone_cbecor_IDX ON whophone (code,"begin","end",capacity,operator,region);


