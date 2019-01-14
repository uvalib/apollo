-- dates like 0/0/YYYY -> YYYY
update nodes set `value` = REPLACE(`value`,'0/0/', '') where node_type_id=28 and value  like '0/0/%';

-- update slash formated dates with unknown DAY
update nodes set `value` = "1960-02-uu" where node_type_id=28 and `value` = "2/0/1960";
update nodes set `value` = "1960-07-uu" where node_type_id=28 and `value` = "7/0/1960";
update nodes set `value` = "1969-05-uu" where node_type_id=28 and `value` = "5/0/1969";
update nodes set `value` = "1960-04-uu" where node_type_id=28 and `value` = "4/0/1960";
update nodes set `value` = "1963-01-uu" where node_type_id=28 and `value` = "1/0/1963";
update nodes set `value` = "1958-04-uu" where node_type_id=28 and `value` = "4/0/1958";
update nodes set `value` = "1960-06-uu" where node_type_id=28 and `value` = "6/0/1960";
update nodes set `value` = "1966-06-uu" where node_type_id=28 and `value` = "6/0/1966";
update nodes set `value` = "1956-02-uu" where node_type_id=28 and `value` = "2/0/1956";
update nodes set `value` = "1956-09-uu" where node_type_id=28 and `value` = "9/0/1956";
update nodes set `value` = "1961-03-uu" where node_type_id=28 and `value` = "3/0/1961";