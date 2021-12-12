DO $$
DECLARE
    v_u_1 UUID;
    v_u_2 UUID;
    
    v_r_superadmin BIGINT;
    v_r_admin BIGINT;
    v_r_user BIGINT;
    v_r_service BIGINT;

BEGIN
INSERT INTO public.users ("username", "password") VALUES ('jochum', '$argon2id$v=19$m=131072,t=4,p=4$1YCaneX7zJb9yeRB4Ej0pw$pLT8wTGpyGQ2CkMPYvsKGzgpd6Wr+S3+BBrzTUEc3cY') RETURNING id INTO v_u_1;
INSERT INTO public.users ("username", "password") VALUES ('pastdue', '$argon2id$v=19$m=131072,t=4,p=4$9yG9qjmNUTslQyZbTe6Knw$GszCEPgCYfJnuxkufTRLI2E0M+s4vr0dJLx05fU2Cps') RETURNING id INTO v_u_2;

INSERT INTO public.roles (name) VALUES ('superadmin') returning id INTO v_r_superadmin;
INSERT INTO public.roles (name) VALUES ('admin') returning id INTO v_r_admin;
INSERT INTO public.roles (name) VALUES ('user') returning id INTO v_r_user;
INSERT INTO public.roles (name) VALUES ('service') returning id INTO v_r_service;

INSERT INTO public.users_roles (role_id, user_id) VALUES (v_r_superadmin, v_u_1), (v_r_admin, v_u_1), (v_r_user, v_u_1);
INSERT INTO public.users_roles (role_id, user_id) VALUES (v_r_superadmin, v_u_2), (v_r_admin, v_u_2),  (v_r_user, v_u_2);
END $$;
