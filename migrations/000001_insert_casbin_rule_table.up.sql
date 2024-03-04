INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES ('p', 'unauthorized', '/v1/register', 'GET');
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES ('p', 'unauthorized', '/v1/swagger/*', 'GET');
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES ('p', 'unauthorized', '/v1/verify', 'GET');
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES ('p', 'unauthorized', '/v1/login', 'POST');
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES ('p', 'user', '/v1/user/update/{id}', 'PUT');
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES ('p', 'user', '/v1/user/delete/{id}', 'DELETE');
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES ('g', 'admin', 'user', '*');
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES ('p', 'admin', '/v1/user/create', 'POST');
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES ('p', 'admin', '/v1/user/{id}', 'GET');
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES ('p', 'user', '/v1/users/{page}/{limit}/{limit}', 'GET');
INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES ('p', 'user', '/v1/user/password/change', 'POST');