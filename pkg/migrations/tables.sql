CREATE TABLE users (
                       id serial primary key,
                       book_id int unique,
                       surname text not null,
                       name text not null,
                       middle_name text not null,
                       birth_date date,
                       student_group text,
                       password bytea,
                       mail text not null,
                       role_level int not null REFERENCES roles(level)
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
);
CREATE TABLE roles (
                       id SERIAL PRIMARY KEY,
                       code TEXT UNIQUE NOT NULL,     -- 'student', 'activist', ...
                       name TEXT NOT NULL,
                       level INT unique NOT NULL CHECK (level > 0)
);
CREATE TABLE refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash BYTEA NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL DEFAULT now() + INTERVAL '30 days',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE TABLE reset_password_tokens (
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    mail TEXT NOT NULL,
    token_hash BYTEA NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id)
);


-- Базовый набор
-- Так же примерный набор прав
-- Уровни с шагом в 10 если вдруг понадобятся промежуточные по правам
INSERT INTO roles (code, name, level) VALUES
                                          -- Просто любой студент
                                          ('student', 'Студент', 10),
                                          -- Имеет стандартные права для игры
                                          -- Не может никак влиять на других пользователей
                                          -- Не может ничего создавать



                                          -- Либо очень доверенный студент, либо, например, преподаватели
                                          -- или вообще люди со стороны, но учавствующие в научной жизни вуза
                                          -- ('supervisor', 'Научный руководитель', 30),
                                          -- Может создавать события/виклики без модерации
                                          -- Может валидировать результаты(начисление балов/опыта) за ИРЛ события

                                          -- Сотрудники деканата
                                          -- ('moderator', 'Модератор', 40),
                                          -- Не может напрямую учавствовать в игре, нет игрового персонажа
                                          -- Может создавать ссылки на регистрацию(не выше supervisor роли)
                                          -- Может управлять ролям юзеров ниже и банить их
                                          -- Область видимоси админ прав - направление, реже - институт

                                          -- Зав. кафедры/РОП
                                          ('admin', 'Администратор', 50),
                                          -- Может создавать ссылки для модераторов и ниже
                                          -- Может управлять ролями и банить модераторов и ниже
                                          -- Может управлять группами внутри своего направления, реже - института
                                          -- Область видимости прав - своё направление, реже - институт

                                          -- Декан/Ректора
                                          -- (условно, вероятно будут замы по научной деятельности или подобное)
                                          -- ('super_admin', 'Супер-администратор', 60),
                                          -- Управление в своей области видимости институтами/направлениями/группами
                                          -- Создание ссылко на все уровни ниже супр-админа
                                          -- Управление ролями и бан пользователей ниже супе-админа
                                          -- Область видимости - институт(для деканов), весь вуз - для ректора


                                          -- Удобная роль для тестирования и полного управления со стороны разработчиков
                                          ('developer', 'Разработчик', 100);
-- Все возможные права что есть выше, как захочется
-- Задаётся ТОЛЬКО через ручное редактировние БД
-- Единственная роль, способная создать ссылку для супер-админа

CREATE UNIQUE INDEX IF NOT EXISTS link_tokens_token_hash_uq
    ON link_tokens (token_hash);