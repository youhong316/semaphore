CREATE TABLE `event` (
  `project_id` int(11) DEFAULT NULL,
  `object_id` int(11) DEFAULT NULL,
  `object_type` varchar(20) DEFAULT '',
  `description` text,
  `created` datetime(6) NOT NULL,
  KEY `project_id` (`project_id`),
  KEY `object_id` (`object_id`),
  KEY `created` (`created`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

alter table task add `created` datetime not null,
	add `start` datetime null,
	add `end` datetime null;