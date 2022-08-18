BEGIN;
CREATE TABLE public.settings
(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    owner_id UUID,
    service varchar(32) COLLATE pg_catalog."default",
    name varchar(32) COLLATE pg_catalog."default" NOT NULL,

    content bytea,

    roles_read varchar(32)[] COLLATE pg_catalog."default",
    roles_update varchar(32)[] COLLATE pg_catalog."default",

    created_at TIMESTAMPTZ DEFAULT Now() NOT NULL,
    updated_at TIMESTAMPTZ NULL,
    deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX osn_idx ON public.settings (owner_id, service, name) WHERE (deleted_at IS NULL);

COMMIT;