CREATE TABLE users (
                       id serial primary key,
                       book_id int unique not null
);
CREATE TABLE students (
                          id serial primary key,
                          book_id int unique not null,
                          surname text not null,
                          name text not null,
                          middle_name text not null,
                          birth_date date,
                          student_group text not null
);
CREATE TABLE link_tokens (
                             book_id     INT NOT NULL REFERENCES students(book_id) ON DELETE CASCADE,
                             token_hash  BYTEA NOT NULL,
                             expires_at  TIMESTAMPTZ NOT NULL,
                             created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
                             PRIMARY KEY (book_id)
)