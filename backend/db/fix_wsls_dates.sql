-- dates like 0/0/YYYY -> YYYY
update nodes set `value` = REPLACE(`value`,'0/0/', '') where node_type_id=28 and value  like '0/0/%';