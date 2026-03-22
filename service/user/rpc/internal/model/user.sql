-- ----------------------------
-- 用户基础表
-- ----------------------------
CREATE TABLE `user` (
                        `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户ID',
                        `username` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '用户名（唯一）',
                        `phone` VARCHAR(16) DEFAULT NULL COMMENT '手机号',
                        `email` VARCHAR(64) DEFAULT NULL COMMENT '邮箱',
                        `password_hash` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '加密密码（bcrypt）',
                        `avatar` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '头像URL',
                        `status` TINYINT NOT NULL DEFAULT '1' COMMENT '状态：1正常，2禁言，3冻结',
                        `is_admin` TINYINT NOT NULL DEFAULT '0' COMMENT '是否管理员：0否，1是',
                        `last_login_ip` VARCHAR(45) NULL DEFAULT '' COMMENT '最后登录IP',
                        `last_login_time` timestamp NULL DEFAULT NULL COMMENT '最后登录时间戳',
                        `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '注册时间',
                        `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                        `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间（软删除）',
                        PRIMARY KEY (`id`),
                        UNIQUE KEY `uk_username` (`username`),
                        UNIQUE KEY `uk_phone` (`phone`),
                        UNIQUE KEY `uk_email` (`email`),
                        KEY `idx_status` (`status`),
                        KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- ----------------------------
-- 用户扩展信息表
-- ----------------------------
CREATE TABLE `user_profile` (
                                `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                                `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
                                `gender` TINYINT NOT NULL DEFAULT '0' COMMENT '性别：0未知，1男，2女',
                                `birthday` DATE DEFAULT NULL COMMENT '生日',
                                `signature` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '个性签名',
                                `location` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '所在地',
                                `website` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '个人网站',
                                `company` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '公司',
                                `job_title` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '职位',
                                `education` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '教育背景',
                                `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                                PRIMARY KEY (`id`),
                                UNIQUE KEY `uk_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户扩展信息表';

-- ----------------------------
-- 用户统计表
-- ----------------------------
CREATE TABLE `user_stat` (
                             `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                             `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
                             `post_count` INT UNSIGNED NOT NULL DEFAULT '0' COMMENT '发帖数',
                             `comment_count` INT UNSIGNED NOT NULL DEFAULT '0' COMMENT '评论数',
                             `follower_count` INT UNSIGNED NOT NULL DEFAULT '0' COMMENT '粉丝数',
                             `following_count` INT UNSIGNED NOT NULL DEFAULT '0' COMMENT '关注数',
                             `like_count` INT UNSIGNED NOT NULL DEFAULT '0' COMMENT '获赞总数',
                             `collection_count` INT UNSIGNED NOT NULL DEFAULT '0' COMMENT '被收藏总数',
                             `last_active_time` BIGINT NOT NULL DEFAULT '0' COMMENT '最后活跃时间',
                             `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                             `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                             PRIMARY KEY (`id`),
                             UNIQUE KEY `uk_user_id` (`user_id`),
                             KEY `idx_last_active` (`last_active_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户统计表';

/*
goctl model mysql ddl -src user.sql -dir ./
goctl model mysql ddl -src user.sql -dir ./internal/model -cache
 */