-- create "history" table
CREATE TABLE
  `history` (
    `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    `mtime` datetime NOT NULL,
    `status` json NULL,
    `definition` json NULL,
    `job_history` text NULL,
    CONSTRAINT `history_job_history` FOREIGN KEY (`job_history`) REFERENCES `job` (`id`) ON DELETE CASCADE
  );

-- create "job" table
CREATE TABLE
  `job` (
    `id` text NOT NULL,
    `stime` datetime NOT NULL,
    `mtime` datetime NOT NULL,
    `client_id` text NOT NULL,
    `definition` json NULL,
    `status` json NOT NULL,
    `group` text NULL,
    `workflow_jobs` integer NULL,
    PRIMARY KEY (`id`),
    CONSTRAINT `job_workflow_jobs` FOREIGN KEY (`workflow_jobs`) REFERENCES `workflow` (`id`) ON DELETE NO ACTION
  );

-- create index "job_stime" to table: "job"
CREATE INDEX `job_stime` ON `job` (`stime`);

-- create index "job_client_id" to table: "job"
CREATE INDEX `job_client_id` ON `job` (`client_id`);

-- create index "job_group" to table: "job"
CREATE INDEX `job_group` ON `job` (`group`);

-- create "tag" table
CREATE TABLE
  `tag` (
    `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    `name` text NOT NULL
  );

-- create index "tag_name" to table: "tag"
CREATE UNIQUE INDEX `tag_name` ON `tag` (`name`);

-- create "workflow" table
CREATE TABLE
  `workflow` (
    `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    `name` text NOT NULL,
    `states` json NOT NULL,
    `transitions` json NOT NULL,
    `groups` json NOT NULL
  );

-- create index "workflow_name_key" to table: "workflow"
CREATE UNIQUE INDEX `workflow_name_key` ON `workflow` (`name`);

-- create "tag_jobs" table
CREATE TABLE
  `tag_jobs` (
    `tag_id` integer NOT NULL,
    `job_id` text NOT NULL,
    PRIMARY KEY (`tag_id`, `job_id`),
    CONSTRAINT `tag_jobs_tag_id` FOREIGN KEY (`tag_id`) REFERENCES `tag` (`id`) ON DELETE CASCADE,
    CONSTRAINT `tag_jobs_job_id` FOREIGN KEY (`job_id`) REFERENCES `job` (`id`) ON DELETE CASCADE
  );

-- added manually, since currently unsupported by entgo
CREATE INDEX idx_status_state ON job (json_extract(status, '$.state'));
