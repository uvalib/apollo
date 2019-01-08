ALTER TABLE publication_history ADD publication varchar(20) not null default 'virgo';
insert into versions (version, created_at) values ('v2', now());