-- get tree
SELECT n.id, n.parent_id, n.sequence, n.pid, n.value, n.created_at, n.updated_at,
   nt.pid, nt.name, nt.controlled_vocab
FROM nodes n
  INNER JOIN node_types nt ON nt.id = n.node_type_id 
WHERE deleted=0 and current=1 and (n.id=1 or ancestry REGEXP '(^.*/|^)1($|/.*)') order by n.id asc;

-- get CHILDREN
SELECT n.id, n.parent_id, n.ancestry, n.sequence, n.pid, n.value, n.created_at, n.updated_at,
   nt.pid, nt.name, nt.controlled_vocab
FROM nodes n
  INNER JOIN node_types nt ON nt.id = n.node_type_id 
WHERE deleted=0 and current=1 and (n.id=11 or ancestry REGEXP '(^.*/|^)11$' and n.value <> "") order by n.id asc;
      
-- get parent collection
SELECT n.id, n.parent_id, n.sequence, n.pid, n.value, n.created_at, n.updated_at,
   nt.pid, nt.name, nt.controlled_vocab
FROM nodes n
  INNER JOIN node_types nt ON nt.id = n.node_type_id 
 WHERE deleted=0 and current=1 and (n.id=1052 or ancestry REGEXP '^1052$' and n.value <> "");
 
select * from nodes where id=411 or parent_id=411 or ancestry like "%/411"  or ancestry like "%/411/%";
select * from nodes where id=442 or parent_id=442 or ancestry like "%/442"  or ancestry like "%/442/%";
select * from nodes where id=582 or parent_id=582 or ancestry like "%/582"  or ancestry like "%/582/%";
