alter table controlled_values drop foreign key controlled_values_ibfk_1;
alter table controlled_values add constraint controlled_values_ibfk_1 FOREIGN KEY (node_type_id) REFERENCES node_types(id) ON DELETE CASCADE;
