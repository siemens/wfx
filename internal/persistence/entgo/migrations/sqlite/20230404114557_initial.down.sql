-- reverse: create "tag_jobs" table
DROP TABLE `tag_jobs`;

-- reverse: create index "workflow_name_key" to table: "workflow"
DROP INDEX `workflow_name_key`;

-- reverse: create "workflow" table
DROP TABLE `workflow`;

-- reverse: create index "tag_name" to table: "tag"
DROP INDEX `tag_name`;

-- reverse: create "tag" table
DROP TABLE `tag`;

-- reverse: create index "job_group" to table: "job"
DROP INDEX `job_group`;

-- reverse: create index "job_client_id" to table: "job"
DROP INDEX `job_client_id`;

-- reverse: create index "job_stime" to table: "job"
DROP INDEX `job_stime`;

-- reverse: create "job" table
DROP TABLE `job`;

-- reverse: create "history" table
DROP TABLE `history`;
