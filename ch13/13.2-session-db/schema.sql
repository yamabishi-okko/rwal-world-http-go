-- schema.sql
CREATE DATABASE IF NOT EXISTS ch13_session_db
  DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_0900_ai_ci;

USE ch13_session_db;

CREATE TABLE IF NOT EXISTS sessions (
  sid        CHAR(32) PRIMARY KEY,         -- ランダムID（32桁）
  data       JSON NOT NULL,                -- 例: {"counter": 1}
  expires_at DATETIME NOT NULL,            -- 有効期限
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);

-- 期限切れクリーン用のインデックス（任意）
CREATE INDEX IF NOT EXISTS idx_sessions_expires
  ON sessions (expires_at);
