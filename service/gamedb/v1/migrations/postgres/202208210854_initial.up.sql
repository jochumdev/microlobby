
BEGIN;

CREATE TABLE public.games
(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    description varchar(64) COLLATE pg_catalog."default" NOT NULL,
    map varchar(40) COLLATE pg_catalog."default" NOT NULL,
    host_ip varchar(40) COLLATE pg_catalog."default" NOT NULL,
    port SMALLINT NOT NULL,
    max_players INTEGER NOT NULL,
    version varchar(64) COLLATE pg_catalog."default" NOT NULL,
    ver_major INTEGER NOT NULL,
    ver_minor INTEGER NOT NULL,
    is_pure BOOL DEFAULT True,
    is_private BOOL default False,
    lobby_version SMALLINT NOT NULL,

    v3_game_id INTEGER NULL,
    mods varchar(64)[] COLLATE pg_catalog."default" NULL,

    created_at TIMESTAMPTZ DEFAULT Now() NOT NULL,
    updated_at TIMESTAMPTZ NULL,
    deleted_at TIMESTAMPTZ NULL
);

CREATE TABLE public.game_players
(
    id BIGSERIAL PRIMARY KEY,
    game_id UUID NOT NULL, 
    uuid UUID NOT NULL,
    name varchar(64) COLLATE pg_catalog."default" NOT NULL,
    ip_address varchar(40) COLLATE pg_catalog."default" NOT NULL,
    is_host BOOL NOT NULL,

    created_at TIMESTAMPTZ DEFAULT Now() NOT NULL,
    updated_at TIMESTAMPTZ NULL,
    deleted_at TIMESTAMPTZ NULL,

    UNIQUE(game_id, uuid),
    FOREIGN KEY(game_id) REFERENCES public.games(id) ON DELETE CASCADE
);
CREATE INDEX game_id_idx ON public.game_players (game_id) WHERE (deleted_at IS NULL);

COMMIT;