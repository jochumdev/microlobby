CREATE TABLE public.roles
(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v1mc(),
    name character varying(32) COLLATE pg_catalog."default" NOT NULL,

    created_at TIMESTAMPTZ DEFAULT Now(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);

CREATE TABLE public.users
(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v1mc(),
    username character varying(255) COLLATE pg_catalog."default" NOT NULL,
    password character varying(255) COLLATE pg_catalog."default",
    email character varying(255) COLLATE pg_catalog."default" NOT NULL,

    created_at TIMESTAMPTZ DEFAULT Now(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);

CREATE TABLE public.users_roles
(
    role_id UUID NOT NULL,
    user_id UUID NOT NULL,

    UNIQUE(role_id, user_id),
    FOREIGN KEY(role_id) REFERENCES public.roles(id) ON DELETE CASCADE,
    FOREIGN KEY(user_id) REFERENCES public.users(id) ON DELETE CASCADE
);
CREATE INDEX users_roles_role_id_idx ON users_roles (role_id);
CREATE INDEX users_roles_user_id_idx ON users_roles (user_id);
