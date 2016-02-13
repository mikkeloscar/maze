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
,owner         TEXT
,name          TEXT
,source_owner  TEXT
,source_name   TEXT
,path          TEXT
,hash          TEXT
,last_check    DATETIME

,UNIQUE(user, name)
);
