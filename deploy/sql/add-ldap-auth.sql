-- LDAP 登录迁移（MySQL 5.7+）
-- 若列已存在会报错，可忽略对应语句后继续执行。

ALTER TABLE `users` ADD COLUMN `auth_source` varchar(20) DEFAULT 'local';
UPDATE `users` SET `auth_source` = 'local' WHERE `auth_source` IS NULL OR `auth_source` = '';
ALTER TABLE `users` MODIFY COLUMN `phone` varchar(30) DEFAULT NULL;
CREATE INDEX `idx_users_auth_source` ON `users` (`auth_source`);
