BEGIN;

CREATE TABLE public.roles
(
    id bigserial PRIMARY KEY,
    name varchar(32) COLLATE pg_catalog."default" NOT NULL,

    created_at TIMESTAMPTZ DEFAULT Now() NOT NULL,
    updated_at TIMESTAMPTZ NULL,
    deleted_at TIMESTAMPTZ NULL
);
CREATE UNIQUE INDEX name_idx ON public.roles (name) WHERE (deleted_at IS NULL);

CREATE TABLE public.users
(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username varchar(255) COLLATE pg_catalog."default" NOT NULL,
    password varchar(255) COLLATE pg_catalog."default" NOT NULL,
    email varchar(255) COLLATE pg_catalog."default" NOT NULL,

    created_at TIMESTAMPTZ DEFAULT Now() NOT NULL,
    updated_at TIMESTAMPTZ NULL,
    deleted_at TIMESTAMPTZ NULL
);
CREATE UNIQUE INDEX username_idx ON public.users (username) WHERE (deleted_at IS NULL);
CREATE UNIQUE INDEX email_idx ON public.users (email) WHERE (deleted_at IS NULL);

CREATE TABLE public.users_roles
(
    user_id UUID NOT NULL,
    role_id BIGINT NOT NULL,

    UNIQUE(role_id, user_id),
    FOREIGN KEY(user_id) REFERENCES public.users(id) ON DELETE CASCADE,
    FOREIGN KEY(role_id) REFERENCES public.roles(id) ON DELETE CASCADE
);
CREATE INDEX user_id_idx ON users_roles (user_id);
CREATE INDEX role_id_idx ON users_roles (role_id);

INSERT INTO roles (name) VALUES ('service'), ('user'), ('admin'), ('superadmin');

INSERT INTO users (id, username, password, email) VALUES ('2e4d8ed5-934d-4cd2-84fb-bd650d3a1ded', 'admin', '$argon2id$v=19$m=131072,t=4,p=4$sMaZvvQn2uWrISQICSbBqQ$L9tNlTTs4ldx0Ry+8Ctu8trSN27Q5iY68iWLjtprOfY', 'admin@wz2100.net');
INSERT INTO users_roles (user_id, role_id) VALUES ('2e4d8ed5-934d-4cd2-84fb-bd650d3a1ded', 2), ('2e4d8ed5-934d-4cd2-84fb-bd650d3a1ded', 3), ('2e4d8ed5-934d-4cd2-84fb-bd650d3a1ded', 4);
COMMIT;