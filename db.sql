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

-- ----------------------------
-- 话题表
-- ----------------------------
CREATE TABLE `topic` (
                         `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '话题ID',
                         `name` VARCHAR(255) NOT NULL COMMENT '话题名',
                         `quote_num` BIGINT NOT NULL DEFAULT '0' COMMENT '引用数',
                         `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                         `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                         `deleted_at` timestamp NULL DEFAULT NULL,
                         PRIMARY KEY (`id`),
                         UNIQUE KEY `uk_name` (`name`),
                         KEY `idx_quote_num` (`quote_num`),
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='话题表';

-- ----------------------------
-- 动态/文章表
-- ----------------------------
CREATE TABLE `post` (
                        `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '动态ID',
                        `user_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '用户ID',
                        `comment_count` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '评论数',
                        `collection_count` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '收藏数',
                        `upvote_count` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '点赞数',
                        `share_count` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '分享数',
                        `visibility` TINYINT NOT NULL DEFAULT '0' COMMENT '可见性:0私密,10充电,20订阅,50好友,60关注,90公开',
                        `is_top` TINYINT NOT NULL DEFAULT '0' COMMENT '置顶',
                        `is_essence` TINYINT NOT NULL DEFAULT '0' COMMENT '精华',
                        `is_lock` TINYINT NOT NULL DEFAULT '0' COMMENT '锁定',
                        `latest_replied_at` BIGINT NOT NULL DEFAULT '0' COMMENT '最后回复时间',
                        `tags` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '标签(逗号分隔)',
                        `attachment_price` BIGINT NOT NULL DEFAULT '0' COMMENT '附件价格(分)',
                        `ip` VARCHAR(45) NOT NULL DEFAULT '' COMMENT '发布IP',
                        `ip_loc` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'IP所在地',
                        `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                        `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                        `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
                        PRIMARY KEY (`id`),
                        KEY `idx_user_id` (`user_id`),
                        KEY `idx_visibility` (`visibility`),
                        KEY `idx_created_at` (`created_at`),
                        CONSTRAINT `fk_post_user_id` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='动态/文章表';

-- ----------------------------
-- 动态内容表
-- ----------------------------
CREATE TABLE `post_content` (
                                `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                                `post_id` BIGINT UNSIGNED NOT NULL COMMENT '动态ID',
                                `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
                                `content` TEXT NOT NULL COMMENT '内容',
                                `type` TINYINT NOT NULL DEFAULT 2 COMMENT '类型：1标题,2文字,3图片,4视频,5语音,6链接,7附件,8收费资源',
                                `sort` INT NOT NULL DEFAULT 100 COMMENT '排序值',
                                `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                                `deleted_at` timestamp NULL DEFAULT NULL,
                                PRIMARY KEY (`id`),
                                KEY `idx_post_id` (`post_id`),
                                KEY `idx_user_id` (`user_id`),
                                CONSTRAINT `fk_content_post_id` FOREIGN KEY (`post_id`) REFERENCES `post` (`id`) ON DELETE CASCADE,
                                CONSTRAINT `fk_content_user_id` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='动态内容表';

-- ----------------------------
-- 动态收藏表
-- ----------------------------
CREATE TABLE `post_collection` (
                                   `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '收藏ID',
                                   `post_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '动态ID',
                                   `user_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '用户ID',
                                   `collected_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '收藏时间',
                                   PRIMARY KEY (`id`),
                                   UNIQUE KEY `uk_user_post` (`user_id`, `post_id`),
                                   KEY `idx_post_id` (`post_id`),
                                   CONSTRAINT `fk_collection_post_id` FOREIGN KEY (`post_id`) REFERENCES `post` (`id`) ON DELETE CASCADE,
                                   CONSTRAINT `fk_collection_user_id` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='动态收藏表';

-- ----------------------------
-- 动态点赞表
-- ----------------------------
CREATE TABLE `post_star` (
                             `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '点赞ID',
                             `post_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '动态ID',
                             `user_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '用户ID',
                             `liked_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '点赞时间',
                             PRIMARY KEY (`id`),
                             UNIQUE KEY `uk_user_post` (`user_id`, `post_id`),
                             KEY `idx_post_id` (`post_id`),
                             CONSTRAINT `fk_star_post_id` FOREIGN KEY (`post_id`) REFERENCES `post` (`id`) ON DELETE CASCADE,
                             CONSTRAINT `fk_star_user_id` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='动态点赞表';

-- ----------------------------
-- 动态话题关联表
-- ----------------------------
CREATE TABLE `post_topic` (
                              `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                              `post_id` BIGINT UNSIGNED NOT NULL,
                              `topic_id` BIGINT UNSIGNED NOT NULL,
                              `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                              PRIMARY KEY (`id`),
                              UNIQUE KEY `uk_post_topic` (`post_id`, `topic_id`),
                              KEY `idx_topic_id` (`topic_id`),
                              CONSTRAINT `fk_pt_post_id` FOREIGN KEY (`post_id`) REFERENCES `post` (`id`) ON DELETE CASCADE,
                              CONSTRAINT `fk_pt_topic_id` FOREIGN KEY (`topic_id`) REFERENCES `topic` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='动态话题关联表';

-- ----------------------------
-- 评论主表
-- ----------------------------
CREATE TABLE `comment` (
                           `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '评论ID',
                           `post_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '动态ID',
                           `user_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '用户ID',
                           `ip` VARCHAR(45) NOT NULL DEFAULT '' COMMENT 'IP地址',
                           `ip_loc` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'IP所在地',
                           `is_essence` TINYINT NOT NULL DEFAULT 0 COMMENT '是否精选',
                           `reply_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '回复数',
                           `thumbs_up_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '点赞数',
                           `thumbs_down_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '点踩数',
                           `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                           `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                           `deleted_at` timestamp NULL DEFAULT NULL,
                           PRIMARY KEY (`id`),
                           KEY `idx_post_id` (`post_id`),
                           KEY `idx_user_id` (`user_id`),
                           CONSTRAINT `fk_comment_post_id` FOREIGN KEY (`post_id`) REFERENCES `post` (`id`) ON DELETE CASCADE,
                           CONSTRAINT `fk_comment_user_id` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论表';

-- ----------------------------
-- 评论内容表
-- ----------------------------
CREATE TABLE `comment_content` (
                                   `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '内容ID',
                                   `comment_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '评论ID',
                                   `user_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '用户ID',
                                   `content` TEXT NOT NULL COMMENT '内容',
                                   `type` TINYINT NOT NULL DEFAULT '2' COMMENT '类型:1标题,2文字,3图片,4视频,5语音,6链接',
                                   `sort` BIGINT NOT NULL DEFAULT '100' COMMENT '排序值',
                                   `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                   `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                                   `deleted_at` timestamp NULL DEFAULT NULL,
                                   PRIMARY KEY (`id`),
                                   KEY `idx_comment_id` (`comment_id`),
                                   KEY `idx_user_id` (`user_id`),
                                   KEY `idx_type` (`type`),
                                   KEY `idx_sort` (`sort`),
                                   CONSTRAINT `fk_comment_content_comment_id` FOREIGN KEY (`comment_id`) REFERENCES `comment` (`id`) ON DELETE CASCADE,
                                   CONSTRAINT `fk_comment_content_user_id` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论内容表';

-- ----------------------------
-- 评论回复表
-- ----------------------------
CREATE TABLE `comment_reply` (
                                 `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '回复ID',
                                 `comment_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '评论ID',
                                 `user_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '用户ID',
                                 `at_user_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '@用户ID',
                                 `content` TEXT NOT NULL COMMENT '内容',
                                 `ip` VARCHAR(45) NOT NULL DEFAULT '' COMMENT 'IP地址',
                                 `ip_loc` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'IP所在地',
                                 `thumbs_up_count` INT UNSIGNED NOT NULL DEFAULT '0' COMMENT '点赞数',
                                 `thumbs_down_count` INT UNSIGNED NOT NULL DEFAULT '0' COMMENT '点踩数',
                                 `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                 `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                                 `deleted_at` timestamp NULL DEFAULT NULL,
                                 PRIMARY KEY (`id`),
                                 KEY `idx_comment_id` (`comment_id`),
                                 CONSTRAINT `fk_comment_reply_comment_id` FOREIGN KEY (`comment_id`) REFERENCES `comment` (`id`) ON DELETE CASCADE,
                                 CONSTRAINT `fk_comment_reply_user_id` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE,
                                 CONSTRAINT `fk_comment_reply_at_user_id` FOREIGN KEY (`at_user_id`) REFERENCES `user` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论回复表';

-- ----------------------------
-- 评论点赞/点踩记录表
-- ----------------------------
CREATE TABLE `comment_thumbs` (
                                  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'thumbs ID',
                                  `user_id` BIGINT UNSIGNED NOT NULL,
                                  `post_id` BIGINT UNSIGNED NOT NULL COMMENT '动态ID',
                                  `comment_id` BIGINT UNSIGNED NOT NULL COMMENT '评论ID',
                                  `reply_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '回复ID(为空表示对评论操作)',
                                  `comment_type` TINYINT NOT NULL DEFAULT '0' COMMENT '0评论,1回复(冗余)',
                                  `is_thumbs_up` TINYINT NOT NULL DEFAULT '0' COMMENT '是否点赞',
                                  `is_thumbs_down` TINYINT NOT NULL DEFAULT '0' COMMENT '是否点踩',
                                  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                                  `deleted_at` timestamp NULL DEFAULT NULL,
                                  PRIMARY KEY (`id`),
                                  UNIQUE KEY `uk_user_comment_reply` (`user_id`, `comment_id`, `reply_id`),
                                  KEY `idx_user_tweet` (`user_id`, `post_id`),
                                  CONSTRAINT `fk_comment_thumbs_user_id` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE,
                                  CONSTRAINT `fk_comment_thumbs_post_id` FOREIGN KEY (`post_id`) REFERENCES `post` (`id`) ON DELETE CASCADE,
                                  CONSTRAINT `fk_comment_thumbs_comment_id` FOREIGN KEY (`comment_id`) REFERENCES `comment` (`id`) ON DELETE CASCADE,
                                  CONSTRAINT `fk_comment_thumbs_reply_id` FOREIGN KEY (`reply_id`) REFERENCES `comment_reply` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论点赞/点踩记录表';

-- ----------------------------
-- 关注表
-- ----------------------------
CREATE TABLE `following` (
                             `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                             `user_id` BIGINT UNSIGNED NOT NULL,
                             `follow_id` BIGINT UNSIGNED NOT NULL,
                             `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                             `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                             `deleted_at` timestamp NULL DEFAULT NULL,
                             PRIMARY KEY (`id`),
                             UNIQUE KEY `uk_user_follow` (`user_id`, `follow_id`),
                             KEY `idx_follow` (`follow_id`),
                             CONSTRAINT `fk_following_user_id` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE,
                             CONSTRAINT `fk_following_follow_id` FOREIGN KEY (`follow_id`) REFERENCES `user` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='关注表';