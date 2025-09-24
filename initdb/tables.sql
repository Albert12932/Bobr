CREATE TABLE users (
                       id serial primary key,
                       book_id int unique not null,
                       surname text not null,
                       name text not null,
                       middle_name text,
                       birth_date date
)
