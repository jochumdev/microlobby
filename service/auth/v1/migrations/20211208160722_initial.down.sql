BEGIN;

DROP TABLE IF EXISTS public.users;
DROP TABLE IF EXISTS public.roles;
DROP TABLE IF EXISTS public.users_roles;
DROP TABLE IF EXISTS public.tokens;
DROP TABLE IF EXISTS public.users_tokens;

COMMIT;