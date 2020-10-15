START TRANSACTION;

CREATE TABLE IF NOT EXISTS versions (
   id int(11) NOT NULL AUTO_INCREMENT PRIMARY KEY,
   version varchar(255) NOT NULL UNIQUE KEY,
   created_at datetime NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- Create table for Apollo users
--
CREATE TABLE IF NOT EXISTS users (
   id int(11) NOT NULL AUTO_INCREMENT PRIMARY KEY,
   computing_id varchar(255) NOT NULL,
   last_name varchar(255) DEFAULT NULL,
   first_name varchar(255) DEFAULT NULL,
   email varchar(255) NOT NULL,
   created_at datetime NOT NULL,
   updated_at datetime NOT NULL,
   UNIQUE KEY (computing_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- Create controlled vocabulary for node names
--
CREATE TABLE IF NOT EXISTS node_types (
   id int(11) NOT NULL AUTO_INCREMENT PRIMARY KEY,
   pid varchar(255) NOT NULL,
   name varchar(255) NOT NULL,
   controlled_vocab boolean not null default 0,
   validation varchar(255) not null default "",
   container boolean not null default 0,
   UNIQUE KEY (name),
   UNIQUE KEY (pid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- Create controlled vocabulary for node values
--
CREATE TABLE IF NOT EXISTS controlled_values (
   id int(11) NOT NULL AUTO_INCREMENT PRIMARY KEY,
   pid varchar(255) NOT NULL,
   node_type_id int(11) NOT NULL,
   value varchar(255) NOT NULL,
   value_uri varchar(255),
   FOREIGN KEY (node_type_id) REFERENCES node_types(id) ON DELETE CASCADE,
   UNIQUE KEY (pid),
   UNIQUE KEY (value)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- Create the main model for tracking serials metadata; the node
-- Notes: parent_id is always the immediate parent
--        ancestry is a path of ids leading to this node
--           Ex: 1/2/5 - 1 is the root node and 5 is immediate parent (same as parent_id)
--        when a revision is made the following happens:
--          1) copy existing node into a new node
--             a) set current = 0
--             b) set PID to the current PID with a '.#'' suffix (an1.1, an1.2)
--          2) update the original node
--        This keeps ID consistent for nodes. History can be found by finding pids with a suffix:
--          select * from nodes where pid like 'p1.%' order by created_at desc
--
CREATE TABLE IF NOT EXISTS nodes (
   id int(11) NOT NULL AUTO_INCREMENT PRIMARY KEY,
   pid varchar(255) NOT NULL,
   parent_id int(11),
   ancestry varchar(255),
   sequence SMALLINT not null,
   node_type_id int(11) NOT NULL,
   value varchar(512) DEFAULT NULL,
   user_id int(11),
   deleted tinyint(1) NOT NULL DEFAULT 0,
   current tinyint(1) NOT NULL DEFAULT 1,
   created_at datetime NOT NULL,
   updated_at datetime,
   KEY (pid),
   UNIQUE KEY (pid),
   KEY (ancestry),
   FOREIGN KEY (node_type_id) REFERENCES node_types(id) ON DELETE RESTRICT,
   FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS publication_history (
  id int(11) NOT NULL AUTO_INCREMENT PRIMARY KEY,
  node_id int(11),
  user_id int(11),
  published_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (node_id) REFERENCES nodes(id) ON DELETE RESTRICT,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

COMMIT;