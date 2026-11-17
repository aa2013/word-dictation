-- ------------------------------------------------------------
-- 0001_init.sql — SQLite 初始表结构（与 model/models.go 对应）
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `user` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `account` TEXT NOT NULL,
  `user_name` TEXT NOT NULL,
  `password` TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS `session` (
  `id` TEXT PRIMARY KEY,
  `user_id` INTEGER NOT NULL,
  `created_at` TEXT NOT NULL,
  `expires_at` TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS `idx_session_user_id` ON `session` (`user_id`);
CREATE INDEX IF NOT EXISTS `idx_session_expires_at` ON `session` (`expires_at`);

CREATE TABLE IF NOT EXISTS `lib` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `lib_name` TEXT NOT NULL,
  `desc` TEXT NOT NULL,
  `cover` TEXT NOT NULL DEFAULT '/covers/default.jpg',
  `creator` INTEGER NOT NULL,
  `common` NUMERIC NOT NULL DEFAULT 0,
  `create_time` TEXT NOT NULL,
  `update_time` TEXT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS `uk_lib_name_creator` ON `lib` (`lib_name`, `creator`);

CREATE TABLE IF NOT EXISTS `lib_word` (
  `lib_id` INTEGER NOT NULL,
  `word_id` INTEGER NOT NULL,
  PRIMARY KEY (`word_id`, `lib_id`)
);

CREATE TABLE IF NOT EXISTS `user_lib` (
  `user_id` INTEGER NOT NULL,
  `lib_id` INTEGER NOT NULL,
  PRIMARY KEY (`user_id`, `lib_id`)
);

CREATE TABLE IF NOT EXISTS `word` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `word` TEXT NOT NULL,
  `us_symbol` TEXT NULL DEFAULT '',
  `us_symbol_mp3` TEXT NULL,
  `en_symbol` TEXT NULL DEFAULT '',
  `en_symbol_mp3` TEXT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS `uk_word_word` ON `word` (`word`);

CREATE TABLE IF NOT EXISTS `word_explain` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `word_id` INTEGER NOT NULL,
  `lib_id` INTEGER NOT NULL,
  `explanation` TEXT NOT NULL,
  `type` TEXT NOT NULL,
  `is_default` NUMERIC NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS `idx_word_explain_word_lib` ON `word_explain` (`word_id`, `lib_id`);

CREATE TABLE IF NOT EXISTS `word_explain_custom` (
  `user_id` INTEGER NOT NULL,
  `lib_id` INTEGER NOT NULL,
  `word_id` INTEGER NOT NULL,
  `exp_id` INTEGER NOT NULL,
  PRIMARY KEY (`word_id`, `lib_id`, `exp_id`, `user_id`)
);

CREATE TABLE IF NOT EXISTS `plan` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `user_id` INTEGER NOT NULL,
  `lib_name` TEXT NOT NULL,
  `name` TEXT NOT NULL,
  `disorder` NUMERIC NOT NULL DEFAULT 1,
  `ch2en` NUMERIC NOT NULL DEFAULT 0,
  `repeat` INTEGER NOT NULL DEFAULT 1,
  `create_time` TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS `plan_word` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `plan_id` INTEGER NOT NULL,
  `word` TEXT NOT NULL,
  `explain` TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS `idx_plan_word_plan_id` ON `plan_word` (`plan_id`);
