CREATE TABLE `stress_task`
(
    `id`         bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `task_id` bigint NOT NULL COMMENT 'task_id',
    `target_qps` int NOT NULL COMMENT 'target_qps',
    `status` tinyint NOT NULL DEFAULT 0 COMMENT 'status 0:未执行 1:等待执行 2:执行 3:更改QPS 4:等待停止 5:已执行',
    `url`varchar(1024) NOT NULL DEFAULT '' COMMENT 'url',
    `protocol`       varchar(20) NOT NULL DEFAULT '' COMMENT 'protocol',
    `method`       varchar(20) NOT NULL DEFAULT '' COMMENT 'method',
    `headers`       varchar(2048) NOT NULL DEFAULT '' COMMENT 'Headers',
    `body`       varchar(2048) NOT NULL DEFAULT '' COMMENT 'body',
    `verify`       varchar(20) NOT NULL DEFAULT '' COMMENT 'verify',
    `timeout`       int NOT NULL DEFAULT 10 COMMENT 'timeout',
    `use_http2`      boolean NOT NULL DEFAULT false COMMENT 'use_http2',
    `keepalive`      boolean NOT NULL DEFAULT false COMMENT 'keepalive',
    `description`       varchar(200) NOT NULL DEFAULT '' COMMENT 'description',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'task create time',
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'task update time',
    `deleted_at` timestamp NULL DEFAULT NULL COMMENT 'task delete time',
    PRIMARY KEY  (`id`),
    UNIQUE KEY   `task_id` (`task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='stress task info table';