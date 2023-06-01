-- create "workflow" table
CREATE TABLE
  `workflow` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `name` varchar(64) NOT NULL,
    `states` json NOT NULL,
    `transitions` json NOT NULL,
    `groups` json NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `name` (`name`)
  ) CHARSET utf8mb4 COLLATE utf8mb4_bin;

-- create "job" table
CREATE TABLE
  `job` (
    `id` varchar(36) NOT NULL,
    `stime` timestamp(6) NOT NULL,
    `mtime` timestamp(6) NOT NULL,
    `client_id` varchar(255) NOT NULL,
    `definition` json NULL,
    `status` json NOT NULL,
    `group` varchar(255) NULL,
    `workflow_jobs` bigint NULL,
    PRIMARY KEY (`id`),
    INDEX `job_client_id` (`client_id`),
    INDEX `job_group` (`group`),
    INDEX `job_stime` (`stime`),
    INDEX `job_workflow_jobs` (`workflow_jobs`),
    -- added manually, since currently unsupported by entgo
    INDEX idx_status_state (
      (JSON_VALUE(status, '$.state' RETURNING CHAR(64)))
    ),
    CONSTRAINT `job_workflow_jobs` FOREIGN KEY (`workflow_jobs`) REFERENCES `workflow` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
  ) CHARSET utf8mb4 COLLATE utf8mb4_bin;

-- create "history" table
CREATE TABLE
  `history` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `mtime` timestamp(6) NOT NULL,
    `status` json NULL,
    `definition` json NULL,
    `job_history` varchar(36) NULL,
    PRIMARY KEY (`id`),
    INDEX `history_job_history` (`job_history`),
    CONSTRAINT `history_job_history` FOREIGN KEY (`job_history`) REFERENCES `job` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
  ) CHARSET utf8mb4 COLLATE utf8mb4_bin;

-- create "tag" table
CREATE TABLE
  `tag` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `name` varchar(255) NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `tag_name` (`name`)
  ) CHARSET utf8mb4 COLLATE utf8mb4_bin;

-- create "tag_jobs" table
CREATE TABLE
  `tag_jobs` (
    `tag_id` bigint NOT NULL,
    `job_id` varchar(36) NOT NULL,
    PRIMARY KEY (`tag_id`, `job_id`),
    INDEX `tag_jobs_job_id` (`job_id`),
    CONSTRAINT `tag_jobs_job_id` FOREIGN KEY (`job_id`) REFERENCES `job` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
    CONSTRAINT `tag_jobs_tag_id` FOREIGN KEY (`tag_id`) REFERENCES `tag` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
  ) CHARSET utf8mb4 COLLATE utf8mb4_bin;
