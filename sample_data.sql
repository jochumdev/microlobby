DO $$
DECLARE
    v_u_a UUID;
    v_u_b UUID;
    
    v_r_superadmin UUID;
    v_r_admin UUID;
    v_r_user UUID;
    v_r_service UUID;

BEGIN
INSERT INTO public.users (username, email) VALUES ('jochum', 'jochum@wz2100.net') RETURNING id INTO v_u_a;
INSERT INTO public.users (username, email) VALUES ('pastdue', 'pastdue@wz2100.net') RETURNING id INTO v_u_b;

INSERT INTO public.roles (name) VALUES ('superadmin') returning id INTO v_r_superadmin;
INSERT INTO public.roles (name) VALUES ('admin') returning id INTO v_r_admin;
INSERT INTO public.roles (name) VALUES ('user') returning id INTO v_r_user;
INSERT INTO public.roles (name) VALUES ('service') returning id INTO v_r_service;

INSERT INTO public.users_roles (role_id, user_id) VALUES (v_r_superadmin, v_u_a), (v_r_admin, v_u_a), (v_r_user, v_u_a);
INSERT INTO public.users_roles (role_id, user_id) VALUES (v_r_user, v_u_b);
END $$;
