-- MySQL dump 10.13  Distrib 5.7.44, for osx10.18 (x86_64)
--
-- Host: 127.0.0.1    Database: admin
-- ------------------------------------------------------
-- Server version	5.7.44

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `admin_depts`
--

DROP TABLE IF EXISTS `admin_depts`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `admin_depts` (
  `dept_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '部门ID',
  `dept_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '部门名称',
  `parent_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '父部门ID',
  `dept_code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '部门编码',
  `description` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '部门描述',
  `sort_order` int(11) NOT NULL DEFAULT '0' COMMENT '排序',
  `dept_source` tinyint(4) NOT NULL DEFAULT '1' COMMENT '部门来源：1自主创建，2企业微信同步，3其他SSO',
  `external_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '外部系统ID（企业微信部门ID等）',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态：1启用，0禁用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`dept_id`) USING BTREE,
  UNIQUE KEY `uk_dept_code` (`dept_code`) USING BTREE,
  KEY `idx_parent_id` (`parent_id`) USING BTREE,
  KEY `idx_dept_source` (`dept_source`) USING BTREE,
  KEY `idx_external_id` (`external_id`) USING BTREE,
  KEY `idx_status` (`status`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='部门表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `admin_depts`
--

LOCK TABLES `admin_depts` WRITE;
/*!40000 ALTER TABLE `admin_depts` DISABLE KEYS */;
/*!40000 ALTER TABLE `admin_depts` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `admin_permissions`
--

DROP TABLE IF EXISTS `admin_permissions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `admin_permissions` (
  `permission_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '权限ID',
  `parent_id` bigint(20) unsigned NOT NULL DEFAULT '0' COMMENT '父权限ID',
  `permission_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '权限名称',
  `permission_code` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '权限编码',
  `permission_type` tinyint(4) NOT NULL DEFAULT '1' COMMENT '权限类型：1菜单，2按钮，3接口',
  `path` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '路径',
  `icon` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '图标',
  `sort_order` int(11) NOT NULL DEFAULT '0' COMMENT '排序',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态：1启用，0禁用',
  `is_system` tinyint(4) NOT NULL DEFAULT '0' COMMENT '是否系统权限：1是，0否',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`permission_id`) USING BTREE,
  UNIQUE KEY `uk_permission_code` (`permission_code`) USING BTREE,
  KEY `idx_parent_id` (`parent_id`) USING BTREE,
  KEY `idx_permission_type` (`permission_type`) USING BTREE,
  KEY `idx_status` (`status`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=58 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `admin_permissions`
--

LOCK TABLES `admin_permissions` WRITE;
/*!40000 ALTER TABLE `admin_permissions` DISABLE KEYS */;
INSERT INTO `admin_permissions` VALUES (1,0,'系统管理','system',1,'','settings_applications',1,1,1,'2025-11-06 14:13:38','2025-11-06 15:31:26'),(2,1,'用户权限管理','system.account',1,'','settings_applications',1,1,1,'2025-11-06 14:13:38','2025-11-06 15:44:55'),(3,2,'用户管理','system.account.user',1,'page/account/user.html','person',1,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(4,3,'新增用户','system.account.user.add',2,'','',1,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(5,3,'编辑用户','system.account.user.edit',2,'','',2,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(6,3,'删除用户','system.account.user.delete',2,'','',3,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(7,3,'刷新列表','system.account.user.refresh',2,'','',4,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(8,3,'获取用户列表','system.account.user.api.list',3,'/admin/user/list','',1,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(9,3,'获取用户信息','system.account.user.api.info',3,'/admin/user/info','',2,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(10,3,'新增用户接口','system.account.user.api.add',3,'/admin/user/add','',3,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(11,3,'更新用户接口','system.account.user.api.update',3,'/admin/user/update','',4,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(12,3,'删除用户接口','system.account.user.api.delete',3,'/admin/user/delete','',5,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(13,3,'更新用户状态接口','system.account.user.api.update-status',3,'/admin/user/update-status','',6,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(14,3,'获取当前用户信息','system.account.user.api.current-info',3,'/admin/user/current-info','',7,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(15,2,'角色管理','system.account.role',1,'page/account/role.html','group',2,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(16,15,'新增角色','system.account.role.add',2,'','',1,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(17,15,'编辑角色','system.account.role.edit',2,'','',2,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(18,15,'删除角色','system.account.role.delete',2,'','',3,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(19,15,'刷新列表','system.account.role.refresh',2,'','',4,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(20,15,'获取角色列表','system.account.role.api.list',3,'/admin/role/list','',1,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(21,15,'获取角色详情','system.account.role.api.detail',3,'/admin/role/detail','',2,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(22,15,'获取编辑角色信息','system.account.role.api.edit-info',3,'/admin/role/edit-info','',3,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(23,15,'新增角色接口','system.account.role.api.add',3,'/admin/role/add','',4,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(24,15,'更新角色接口','system.account.role.api.update',3,'/admin/role/update','',5,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(25,15,'删除角色接口','system.account.role.api.delete',3,'/admin/role/delete','',6,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(26,15,'更新角色状态接口','system.account.role.api.update-status',3,'/admin/role/update-status','',7,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(27,15,'更新是否系统角色接口','system.account.role.api.update-is-system',3,'/admin/role/update-is-system','',8,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(28,15,'获取所有权限接口','system.account.role.api.get-all-permissions',3,'/admin/role/get-all-permissions','',9,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(29,15,'获取用户角色和权限','system.account.role.api.get-user-role-permissions',3,'/admin/role/get-user-role-permissions','',10,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(30,2,'权限管理','system.account.permission',1,'page/account/permission.html','lock',3,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(31,30,'新增权限','system.account.permission.add',2,'','',1,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(32,30,'编辑权限','system.account.permission.edit',2,'','',2,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(33,30,'删除权限','system.account.permission.delete',2,'','',3,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(34,30,'刷新列表','system.account.permission.refresh',2,'','',4,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(35,30,'获取权限列表','system.account.permission.api.list',3,'/admin/permission/list','',1,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(36,30,'获取编辑权限信息','system.account.permission.api.edit-info',3,'/admin/permission/edit-info','',2,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(37,30,'获取权限树','system.account.permission.api.get-tree',3,'/admin/permission/get-tree','',3,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(38,30,'新增权限接口','system.account.permission.api.add',3,'/admin/permission/add','',4,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(39,30,'更新权限接口','system.account.permission.api.update',3,'/admin/permission/update','',5,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(40,30,'删除权限接口','system.account.permission.api.delete',3,'/admin/permission/delete','',6,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(41,30,'更新权限状态接口','system.account.permission.api.update-status',3,'/admin/permission/update-status','',7,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(42,30,'更新是否系统权限接口','system.account.permission.api.update-is-system',3,'/admin/permission/update-is-system','',8,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(43,1,'后台日志','system.log',1,'','markunread_mailbox',2,1,1,'2025-11-06 14:13:38','2025-11-06 15:31:26'),(44,43,'登录日志','system.log.login',1,'page/log/login.html','history',1,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(45,44,'查询登录日志','system.log.login.query',2,'','',1,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(46,44,'获取登录日志列表','system.log.login.api.list',3,'/admin/login-log/list','',1,1,1,'2025-11-06 14:13:38','2025-11-06 14:13:38'),(47,3,'用户管理列表查询','system.accout.user.query',2,'','',0,1,0,'2025-11-06 16:03:52','2025-11-06 16:03:52'),(48,15,'角色管理列表查询','system.account.role.query',2,'','',0,1,0,'2025-11-06 16:04:53','2025-11-06 16:05:14'),(49,30,'权限管理列表查询','system.account.permission.query',2,'','',1,1,0,'2025-11-06 16:07:37','2025-11-06 16:24:19'),(50,1,'信息管理','system.profile',1,'','history',0,1,1,'2025-11-06 19:04:29','2025-11-06 19:08:40'),(51,50,'个人信息','system.profile.info',1,'page/profile/profile.html','',0,1,1,'2025-11-06 19:06:14','2025-11-06 19:09:00'),(52,51,'个人信息更新','system.profile.info.updateProfile',2,'','',0,1,0,'2025-11-07 11:34:54','2025-11-07 11:34:54'),(53,51,'修改密码','system.profile.info.password',2,'','',0,1,0,'2025-11-07 11:35:43','2025-11-07 11:35:43'),(54,51,'绑定与解绑手机号码','system.profile.info.mobile',2,'','',0,1,0,'2025-11-07 11:36:27','2025-11-07 11:36:27'),(55,51,'修改密码','system.profile.info.password.changePassword',3,'/admin/user/change-password','',0,1,0,'2025-11-07 11:40:36','2025-11-07 11:43:27'),(56,51,'绑定手机号','system.profile.info.bindMobile',3,'/admin/user/bind-mobile','',0,1,0,'2025-11-07 11:41:30','2025-11-07 11:46:28'),(57,51,'解绑手机号','system.profile.info.unbindMobile',3,'/admin/user/unbind-mobile','',0,1,0,'2025-11-07 11:42:21','2025-11-07 12:20:00');
/*!40000 ALTER TABLE `admin_permissions` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `admin_role_permissions`
--

DROP TABLE IF EXISTS `admin_role_permissions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `admin_role_permissions` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `role_id` bigint(20) unsigned NOT NULL COMMENT '角色ID',
  `permission_id` bigint(20) unsigned NOT NULL COMMENT '权限ID',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_role_permission` (`role_id`,`permission_id`) USING BTREE,
  KEY `idx_role_id` (`role_id`) USING BTREE,
  KEY `idx_permission_id` (`permission_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=232 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色权限关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `admin_role_permissions`
--

LOCK TABLES `admin_role_permissions` WRITE;
/*!40000 ALTER TABLE `admin_role_permissions` DISABLE KEYS */;
INSERT INTO `admin_role_permissions` VALUES (130,2,50,'2025-11-07 11:52:58','2025-11-07 11:52:58'),(131,2,51,'2025-11-07 11:52:58','2025-11-07 11:52:58'),(132,2,55,'2025-11-07 11:52:58','2025-11-07 11:52:58'),(133,2,56,'2025-11-07 11:52:58','2025-11-07 11:52:58'),(134,2,57,'2025-11-07 11:52:58','2025-11-07 11:52:58'),(135,2,54,'2025-11-07 11:52:58','2025-11-07 11:52:58'),(136,2,52,'2025-11-07 11:52:58','2025-11-07 11:52:58'),(137,2,53,'2025-11-07 11:52:58','2025-11-07 11:52:58'),(138,2,43,'2025-11-07 11:52:58','2025-11-07 11:52:58'),(139,2,44,'2025-11-07 11:52:58','2025-11-07 11:52:58'),(140,2,46,'2025-11-07 11:52:58','2025-11-07 11:52:58'),(141,2,45,'2025-11-07 11:52:58','2025-11-07 11:52:58'),(153,3,1,'2025-11-07 12:06:48','2025-11-07 12:06:48'),(154,3,50,'2025-11-07 12:06:48','2025-11-07 12:06:48'),(155,3,51,'2025-11-07 12:06:48','2025-11-07 12:06:48');
/*!40000 ALTER TABLE `admin_role_permissions` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `admin_roles`
--

DROP TABLE IF EXISTS `admin_roles`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `admin_roles` (
  `role_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '角色ID',
  `role_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色名称',
  `role_code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色编码',
  `description` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '角色描述',
  `sort_order` int(11) NOT NULL DEFAULT '0' COMMENT '排序',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态：1启用，0禁用',
  `is_system` tinyint(4) NOT NULL DEFAULT '0' COMMENT '是否系统角色：1是，0否',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`role_id`) USING BTREE,
  UNIQUE KEY `uk_role_code` (`role_code`) USING BTREE,
  KEY `idx_status` (`status`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `admin_roles`
--

LOCK TABLES `admin_roles` WRITE;
/*!40000 ALTER TABLE `admin_roles` DISABLE KEYS */;
INSERT INTO `admin_roles` VALUES (1,'超级管理员','super_admin','系统超级管理员，拥有所有权限',0,1,1,'2025-11-06 06:07:38','2025-11-06 06:07:38'),(2,'普通管理员','admin','普通管理员',0,1,1,'2025-11-06 06:07:38','2025-11-07 11:53:02'),(3,'普通用户','user','普通用户',0,1,0,'2025-11-06 06:07:38','2025-11-07 12:06:48');
/*!40000 ALTER TABLE `admin_roles` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `admin_user_depts`
--

DROP TABLE IF EXISTS `admin_user_depts`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `admin_user_depts` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `user_id` bigint(20) unsigned NOT NULL COMMENT '用户ID',
  `dept_id` bigint(20) unsigned NOT NULL COMMENT '部门ID',
  `is_primary` tinyint(4) NOT NULL DEFAULT '0' COMMENT '是否主要部门：1是，0否',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态：1启用，0禁用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_user_dept` (`user_id`,`dept_id`) USING BTREE,
  KEY `idx_user_id` (`user_id`) USING BTREE,
  KEY `idx_dept_id` (`dept_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户部门关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `admin_user_depts`
--

LOCK TABLES `admin_user_depts` WRITE;
/*!40000 ALTER TABLE `admin_user_depts` DISABLE KEYS */;
/*!40000 ALTER TABLE `admin_user_depts` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `admin_user_login_logs`
--

DROP TABLE IF EXISTS `admin_user_login_logs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `admin_user_login_logs` (
  `log_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `user_id` bigint(20) unsigned NOT NULL COMMENT '用户ID',
  `login_type` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '登录方式',
  `login_ip` varchar(45) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '登录IP',
  `user_agent` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '用户代理',
  `login_status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '登录状态：1成功，0失败',
  `failure_reason` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '失败原因',
  `login_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间',
  PRIMARY KEY (`log_id`) USING BTREE,
  KEY `idx_user_id` (`user_id`) USING BTREE,
  KEY `idx_login_time` (`login_time`) USING BTREE,
  KEY `idx_login_status` (`login_status`) USING BTREE,
  KEY `idx_login_type` (`login_type`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户登录日志表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `admin_user_login_logs`
--

LOCK TABLES `admin_user_login_logs` WRITE;
/*!40000 ALTER TABLE `admin_user_login_logs` DISABLE KEYS */;
INSERT INTO `admin_user_login_logs` VALUES (1,1,'logout','::1','Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36',1,'用户主动登出','2025-11-07 15:02:53'),(2,1,'mobile','::1','Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36',1,'','2025-11-07 15:03:53');
/*!40000 ALTER TABLE `admin_user_login_logs` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `admin_user_logins`
--

DROP TABLE IF EXISTS `admin_user_logins`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `admin_user_logins` (
  `login_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '登录方式ID',
  `user_id` bigint(20) unsigned NOT NULL COMMENT '用户ID',
  `login_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '登录类型：password、sms、oauth:provider',
  `login_value` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '登录标识（手机号/用户名/unionid或openid）',
  `login_credential` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '登录凭证（密码hash；不得存access_token）',
  `salt` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '密码盐值',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态：1启用，0禁用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`login_id`) USING BTREE,
  UNIQUE KEY `uk_login_type_value` (`login_type`,`login_value`) USING BTREE,
  UNIQUE KEY `uk_user_login_type` (`user_id`,`login_type`) USING BTREE,
  KEY `idx_user_id` (`user_id`) USING BTREE,
  KEY `idx_login_type` (`login_type`) USING BTREE,
  KEY `idx_status` (`status`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户登录方式表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `admin_user_logins`
--

LOCK TABLES `admin_user_logins` WRITE;
/*!40000 ALTER TABLE `admin_user_logins` DISABLE KEYS */;
INSERT INTO `admin_user_logins` VALUES (1,1,'mobile','13013013000','','',1,'2025-11-07 15:03:52','2025-11-07 15:03:52');
/*!40000 ALTER TABLE `admin_user_logins` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `admin_user_roles`
--

DROP TABLE IF EXISTS `admin_user_roles`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `admin_user_roles` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `user_id` bigint(20) unsigned NOT NULL COMMENT '用户ID',
  `role_id` bigint(20) unsigned NOT NULL COMMENT '角色ID',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_user_role` (`user_id`,`role_id`) USING BTREE,
  KEY `idx_user_id` (`user_id`) USING BTREE,
  KEY `idx_role_id` (`role_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `admin_user_roles`
--

LOCK TABLES `admin_user_roles` WRITE;
/*!40000 ALTER TABLE `admin_user_roles` DISABLE KEYS */;
INSERT INTO `admin_user_roles` VALUES (10,1,1,'2025-11-06 14:38:12','2025-11-06 14:38:12'),(11,2,3,'2025-11-07 12:19:05','2025-11-07 12:19:05');
/*!40000 ALTER TABLE `admin_user_roles` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `admin_users`
--

DROP TABLE IF EXISTS `admin_users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `admin_users` (
  `user_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `username` varchar(60) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '用户名（必填，等于手机号）',
  `realname` varchar(60) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '真实姓名',
  `nickname` varchar(60) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '昵称',
  `avatar` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '头像URL',
  `mobile` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '手机号码（可为空；第三方未返回手机号时为NULL）',
  `email` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '邮箱地址',
  `gender` tinyint(4) NOT NULL DEFAULT '0' COMMENT '性别：0未知，1男，2女',
  `birthday` date DEFAULT NULL COMMENT '生日',
  `position` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '职位',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态：1启用，0禁用，-1删除',
  `user_source` tinyint(4) NOT NULL DEFAULT '1' COMMENT '用户来源：1自主注册，2第三方登录, 3. 管理员创建',
  `last_login_time` datetime DEFAULT NULL COMMENT '最后登录时间',
  `last_login_ip` varchar(45) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '最后登录IP',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`user_id`) USING BTREE,
  UNIQUE KEY `uk_username` (`username`) USING BTREE,
  UNIQUE KEY `uk_mobile` (`mobile`) USING BTREE,
  UNIQUE KEY `uk_email` (`email`) USING BTREE,
  KEY `idx_status` (`status`) USING BTREE,
  KEY `idx_user_source` (`user_source`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户基础信息表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `admin_users`
--

LOCK TABLES `admin_users` WRITE;
/*!40000 ALTER TABLE `admin_users` DISABLE KEYS */;
INSERT INTO `admin_users` VALUES (1,'13013013000','Taurus','stones','','13013013000','demo@taurus.com',1,'1900-01-01','',1,1,'2025-11-07 15:03:53','::1','2025-11-07 15:03:52','2025-11-07 15:22:18');
/*!40000 ALTER TABLE `admin_users` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Dumping events for database 'admin'
--

--
-- Dumping routines for database 'admin'
--
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2025-11-07 15:28:55
