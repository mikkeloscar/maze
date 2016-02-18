-- +migrate Up

CREATE TABLE users (
 id     INTEGER PRIMARY KEY AUTOINCREMENT
,login  TEXT
,token  TEXT
,admin  BOOLEAN
,hash   TEXT

,UNIQUE(login)
);

CREATE TABLE repos (
 id            INTEGER PRIMARY KEY AUTOINCREMENT
,user_id       INTEGER
,private       BOOLEAN
,owner         TEXT
,name          TEXT
,source_owner  TEXT
,source_name   TEXT
,hash          TEXT
,last_check    DATETIME

,UNIQUE(owner, name)
);
