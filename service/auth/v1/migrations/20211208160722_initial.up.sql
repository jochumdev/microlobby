BEGIN;

CREATE TABLE public.roles
(
    id bigserial PRIMARY KEY,
    name varchar(32) COLLATE pg_catalog."default" NOT NULL,

    created_at TIMESTAMPTZ DEFAULT Now(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);

CREATE TABLE public.users
(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v1mc(),
    username varchar(255) COLLATE pg_catalog."default" NOT NULL,
    password varchar(255) COLLATE pg_catalog."default" NOT NULL,
    email varchar(255) COLLATE pg_catalog."default" NOT NULL,

    created_at TIMESTAMPTZ DEFAULT Now(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);

CREATE TABLE public.users_roles
(
    user_id UUID NOT NULL,
    role_id BIGINT NOT NULL,

    UNIQUE(role_id, user_id),
    FOREIGN KEY(user_id) REFERENCES public.users(id) ON DELETE CASCADE,
    FOREIGN KEY(role_id) REFERENCES public.roles(id) ON DELETE CASCADE
);
CREATE INDEX users_roles_user_id_idx ON users_roles (user_id);
CREATE INDEX users_roles_role_id_idx ON users_roles (role_id);

COMMIT;