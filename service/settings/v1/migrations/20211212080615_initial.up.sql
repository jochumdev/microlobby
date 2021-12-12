CREATE TABLE public.settings
(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v1mc(),

    owner_id UUID,
    service varchar(32) COLLATE pg_catalog."default",
    name varchar(32) COLLATE pg_catalog."default" NOT NULL,

    content TEXT,

    roles_read varchar(32)[] COLLATE pg_catalog."default",
    roles_update varchar(32)[] COLLATE pg_catalog."default",

    created_at TIMESTAMPTZ DEFAULT Now(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);