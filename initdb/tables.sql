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
)