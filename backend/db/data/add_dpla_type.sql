insert into node_types(pid,name,controlled_vocab,container, validation) values
   ("uva-ant29", "dpla", 0,0,"");

--- raw insert to add DPLA nodes to each collection
INSERT INTO `nodes` (`id`, `pid`, `parent_id`, `ancestry`, `sequence`, `node_type_id`, `value`, `user_id`, `deleted`, `current`, `created_at`, `updated_at`)
VALUES
	(321361, 'uva-an321361', 1, '1', 10, 29, '0', 1, 0, 1, '2019-01-07 11:06:03', NULL),
	(321362, 'uva-an321362', 118, '118', 31, 29, '0', 1, 0, 1, '2019-01-07 11:24:08', NULL),
	(321363, 'uva-an321363', 1054, '1054', 79, 29, '0', 1, 0, 1, '2019-01-07 11:25:01', NULL),
	(321364, 'uva-an321364', 109873, '109873', 26, 29, '1', 1, 0, 1, '2019-01-07 11:25:01', NULL);
