
/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Create table for Apollo users
--
DROP TABLE IF EXISTS versions;
CREATE TABLE versions (
   id int(11) NOT NULL AUTO_INCREMENT PRIMARY KEY,
   version varchar(255) NOT NULL UNIQUE KEY,
   created_at datetime NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
insert into versions(version, created_at) values ("v1", NOW());

--
-- Create table for Apollo users
--
DROP TABLE IF EXISTS users;
CREATE TABLE users (
   id int(11) NOT NULL AUTO_INCREMENT PRIMARY KEY,
   computing_id varchar(255) NOT NULL,
   last_name varchar(255) DEFAULT NULL,
   first_name varchar(255) DEFAULT NULL,
   email varchar(255) NOT NULL,
   created_at datetime NOT NULL,
   updated_at datetime NOT NULL,
   UNIQUE KEY (computing_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- Add starter user
insert into users(computing_id, last_name,first_name,email,created_at,updated_at)
   values ("lf6f", "Foster", "Lou", "lf6f@virginia.edu", NOW(), NOW()),
	        ('md5wz', 'Mike', 'Durbin', 'md5wz@virginia.edu', NOW(), NOW() );

--
-- Create controlled vocabulary for node names
--
DROP TABLE IF EXISTS node_names;
DROP TABLE IF EXISTS node_types;
CREATE TABLE node_types (
   id int(11) NOT NULL AUTO_INCREMENT PRIMARY KEY,
   pid varchar(255) NOT NULL,
   name varchar(255) NOT NULL,
   controlled_vocab boolean not null default 0,
   validation varchar(255) not null default "",
   container boolean not null default 0,
   UNIQUE KEY (name),
   UNIQUE KEY (pid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- add some starter types
insert into node_types(pid,name,controlled_vocab,container, validation) values
   ("uva-ant1", "collection", 0,1,null), ("uva-ant2", "title", 0,0, null),
   ("uva-ant3", "volume", 0,1, null), ("uva-ant4", "issue", 0,1, null),
	 ('uva-ant5', 'externalPID', 0,0, null), ('uva-ant6', 'digitalObject', 0,0, null),
	 ('uva-ant7', 'year', 0,1, "^\d{4}$"), ('uva-ant8', 'month', 0,1, null),
   ('uva-ant9', 'barcode', 0,0, null), ('uva-ant10', 'catalogKey', 0,0, null),
   ('uva-ant11', 'useRights', 1,0, null);


--
-- Create controlled vocabulary for node values
--
DROP TABLE IF EXISTS controlled_values;
CREATE TABLE controlled_values (
   id int(11) NOT NULL AUTO_INCREMENT PRIMARY KEY,
   pid varchar(255) NOT NULL,
   node_type_id int(11) NOT NULL,
   value varchar(255) NOT NULL,
   value_uri varchar(255),
   FOREIGN KEY (node_type_id) REFERENCES node_names(id) ON DELETE CASCADE,
   UNIQUE KEY (pid),
   UNIQUE KEY (value)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
INSERT INTO controlled_values (pid, node_type_id, value, value_uri)
VALUES
	('uva-acv1', 11, 'Copyright Not Evaluated', 'http://rightsstatements.org/vocab/CNE/1.0/'),
	('uva-acv2', 11, 'No Known Copyright', 'http://rightsstatements.org/vocab/NKC/1.0/'),
	('uva-acv3', 11, 'In Copyright', 'http://rightsstatements.org/vocab/InC/1.0/'),
	('uva-acv4', 11, 'In Copyright Educational Use Permitted', 'http://rightsstatements.org/vocab/InC-EDU/1.0/'),
	('uva-acv5', 11, 'In Copyright Non-Commercial Use Permitted', 'http://rightsstatements.org/vocab/InC-NC/1.0/'),
	('uva-acv6', 11, 'In Copyright Rights Holder Unlocatable', 'http://rightsstatements.org/vocab/InC-RUU/1.0/'),
	('uva-acv9', 11, 'No Copyright Other Known Legal Restrictions', 'http://rightsstatements.org/vocab/NoC-OKLR/1.0/'),
	('uva-acv10', 11, 'No Copyright United States', 'http://rightsstatements.org/vocab/NoC-US/1.0/'),
	('uva-acv11', 11, 'Copyright Undetermined', 'http://rightsstatements.org/vocab/UND/1.0/');



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
DROP TABLE IF EXISTS nodes;
CREATE TABLE nodes (
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

/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
