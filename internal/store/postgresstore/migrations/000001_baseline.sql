-- +goose Up
create table hellos (
  id UUID primary key,
  name text not null
);

CREATE INDEX hello_idx_id ON hellos (id);
CREATE INDEX hello_idx_name ON hellos (name);

-- +goose Down
drop table hellos;
drop index hello_idx_id;
drop index hello_idx_name;
