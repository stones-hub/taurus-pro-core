-- 设置密码为native_password, mysql 8.0 默认使用 caching_sha2_password
ALTER USER 'apps'@'%' IDENTIFIED WITH caching_sha2_password BY 'apps_password';

-- 设置权限
GRANT ALL PRIVILEGES ON admin.* TO 'apps'@'%';

-- 刷新权限
FLUSH PRIVILEGES;
