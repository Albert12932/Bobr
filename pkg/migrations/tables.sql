CREATE TABLE IF NOT EXISTS roles (
                                     id SERIAL PRIMARY KEY,
                                     code TEXT UNIQUE NOT NULL,     -- 'student', 'activist', ...
                                     name TEXT NOT NULL,
                                     level INT unique NOT NULL CHECK (level > 0)
);
CREATE TABLE IF NOT EXISTS users (
                       id serial primary key,
                       book_id int unique,
                       surname text not null,
                       name text not null,
                       middle_name text not null,
                       birth_date date,
                       student_group text,
                       password bytea,
                       email text not null,
                       role_level int not null REFERENCES roles(level)
);
CREATE TABLE IF NOT EXISTS students (
                          id serial primary key,
                          book_id int unique not null,
                          surname text not null,
                          name text not null,
                          middle_name text not null,
                          birth_date date,
                          student_group text not null
);
CREATE TABLE IF NOT EXISTS link_tokens (
                             book_id     INT NOT NULL REFERENCES students(book_id) ON DELETE CASCADE,
                             token_hash  BYTEA NOT NULL,
                             expires_at  TIMESTAMPTZ NOT NULL,
                             created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
                             PRIMARY KEY (book_id)
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash BYTEA NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL DEFAULT now() + INTERVAL '30 days',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS reset_password_tokens (
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    token_hash BYTEA NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id)
);
CREATE TABLE IF NOT EXISTS institutes (
                                          id serial primary key,
                                          name text unique not null
);
CREATE TABLE IF NOT EXISTS studies (
                                       id serial primary key,
                                       name text unique not null,
                                       institute_id int references institutes(id) ON DELETE CASCADE,
                                       FOREIGN KEY (institute_id) REFERENCES institutes(id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS student_groups (
    id serial primary key,
    name text unique not null,
    studies_id int references studies(id) on DELETE CASCADE,
    FOREIGN KEY (studies_id) references studies(id) ON DELETE CASCADE

);

CREATE TABLE IF NOT EXISTS events_types (
    id serial primary key,
    code int unique not null,
    name text not null
);


CREATE TABLE IF NOT EXISTS events (
                                      id serial primary key,
                                      event_type_code int default 0 references events_types(code) on DELETE SET DEFAULT,
                                      title text unique not null,
                                      description text default 'Empty description',
                                      points int default 100,
                                      icon_url text default 'https://09edcbd14ce2e9c5981946024728da15.bckt.ru/testIcons/star.webp',
                                      event_date timestamptz default '1970-01-01T00:00:00Z',
                                      created_at timestamptz default now()
);
CREATE TABLE IF NOT EXISTS completed_events (
    user_id int references users(id) on DELETE CASCADE,
    event_id int references events(id) on DELETE CASCADE,
    completed_at timestamptz default now(),
    PRIMARY KEY (user_id, event_id)
);

INSERT INTO roles (code, name, level) VALUES
                                          ('student', 'Студент', 10),
                                          ('admin', 'Администратор', 50),
                                          ('developer', 'Разработчик', 100);

INSERT into events_types (code, name) VALUES
                                          (1, 'Хакатон'),
                                          (2, 'Статья'),
                                          (3, 'Олимпиада'),
                                          (4, 'Проект'),
                                          (0, 'Неизвестно');

CREATE UNIQUE INDEX IF NOT EXISTS link_tokens_token_hash_uq
    ON link_tokens (token_hash);

