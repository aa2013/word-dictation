-- ------------------------------------------------------------
-- 0003_setting.sql — 系统键值配置表与初始化标记
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `setting` (
  `key` TEXT PRIMARY KEY,
  `value` TEXT NOT NULL
);
INSERT OR IGNORE INTO `setting` (`key`, `value`) VALUES ('sys_initialized', '0');
