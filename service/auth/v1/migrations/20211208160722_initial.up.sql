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
    FOREIGN KEY(user_id) REFERENCES public.users(id),
    FOREIGN KEY(role_id) REFERENCES public.roles(id)
);
CREATE INDEX user_id_idx ON users_roles (user_id);
CREATE INDEX role_id_idx ON users_roles (role_id);

INSERT INTO roles (name) VALUES ('service');
INSERT INTO roles (name) VALUES ('user');
INSERT INTO roles (name) VALUES ('admin');
INSERT INTO roles (name) VALUES ('superadmin');

COMMIT;