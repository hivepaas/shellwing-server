-- The data encryption key every seeded secret is sealed with
INSERT INTO encryption_keys (id, wrapped_key, is_active, created_at, updated_at)
VALUES ('01JZZZZZZZZZZZZZZZZZZZZZZZ', 'hpsalt:jdJ9srZ2VIHDJQ== QtiXpX9IfIqe7S1gOfmxo51sgkMOUFdamM2yo7NIoKVqNtF2TU8pAKk5W+6fJXVpMutD46pS7CCqx1Q8XYGW0CbtPIfRszgh', TRUE, NOW(), NOW());

-- Users
INSERT INTO users (id, username, email, role, status, full_name, position, security_option, totp_secret, password, created_at, updated_at)
VALUES ('01JAB9XED0GTXBSQDFVYAJ8WA1', 'tiendc', 'tiendc@gmail.com', 'admin', 'active', 'Tien DC', 'manager', 'password-only', NULL,
        'VanDSYUU/sHz1w== 3vTE3ZLtpeM23UGllSazQh88/I06G5VrNzHZtBbGjFQ=',
        '2025-10-01 00:00:00', '2025-10-01 00:00:00'),
       ('01JAB9XED0GTXBSQDFVYAJ8WA2', 'member1', 'member1@domain.name', 'member', 'active', 'Member 1', 'devops', 'password-only', NULL,
        'VanDSYUU/sHz1w== 3vTE3ZLtpeM23UGllSazQh88/I06G5VrNzHZtBbGjFQ=',
        '2025-10-01 00:00:00', '2025-10-01 00:00:00'),
       ('01JAB9XED0GTXBSQDFVYAJ8WA3', 'member2', 'member2@domain.name', 'member', 'active', 'Member 2', 'developer', 'password-only', NULL,
        'VanDSYUU/sHz1w== 3vTE3ZLtpeM23UGllSazQh88/I06G5VrNzHZtBbGjFQ=',
        '2025-10-01 00:00:00', '2025-10-01 00:00:00'),
       ('01JAB9XED0GTXBSQDFVYAJ8WA4', 'member3', 'member3@domain.name', 'member', 'active', 'Member 3', NULL, 'password-2fa', 'AAAAAAAAAAAAAAAAAAAA',
        'VanDSYUU/sHz1w== 3vTE3ZLtpeM23UGllSazQh88/I06G5VrNzHZtBbGjFQ=',
        '2025-10-01 00:00:00', '2025-10-01 00:00:00'),
       ('01JAB9XED0GTXBSQDFVYAJ8WA5', 'member4', 'member4@domain.name', 'member', 'pending', 'Member 4', NULL, 'password-2fa', NULL,
        'VanDSYUU/sHz1w== 3vTE3ZLtpeM23UGllSazQh88/I06G5VrNzHZtBbGjFQ=',
        '2025-10-01 00:00:00', '2025-10-01 00:00:00'),
       ('01JAB9XED0GTXBSQDFVYAJ8WA6', 'member5', 'member5@domain.name', 'member', 'pending', 'Member 5', NULL, 'password-2fa', NULL,
        'VanDSYUU/sHz1w== 3vTE3ZLtpeM23UGllSazQh88/I06G5VrNzHZtBbGjFQ=',
        '2025-10-01 00:00:00', '2025-10-01 00:00:00'),
       ('01JAB9XED0GTXBSQDFVYAJ8WA7', 'admin1', 'admin1@domain.name', 'admin', 'active', 'Admin 1', 'developer', 'password-only', NULL,
        'VanDSYUU/sHz1w== 3vTE3ZLtpeM23UGllSazQh88/I06G5VrNzHZtBbGjFQ=',
        '2025-10-01 00:00:00', '2025-10-01 00:00:00')
ON CONFLICT DO NOTHING;

INSERT INTO acl_permissions (subj_type, subj_id, res_type, res_id, p_read, p_exec, p_write, p_del, created_at, updated_at)
VALUES ('user', '01JAB9XED0GTXBSQDFVYAJ8WA2', 'module', 'mod::user', true, true, true, false,
        '2025-10-01 00:00:00', '2025-10-01 00:00:00'),
       ('user', '01JAB9XED0GTXBSQDFVYAJ8WA2', 'module', 'mod::settings', true, true, true, true,
        '2025-10-01 00:00:00', '2025-10-01 00:00:00')
ON CONFLICT DO NOTHING;

-- Settings: Providers
INSERT INTO settings (id, scope, object_id, type, kind, name, status, inheritable, is_default, data, created_at, updated_at)
VALUES ('01JAB9XED0GTXBSQDFVYAJ8WE1', '', NULL, 'oauth', 'github', 'Github', 'active', true, true,
        '{"org": "hivepaas-test", "clientId": "Ov23lirztQpWxZTKNcTQ", "clientSecret": "hpenc:l1IFVb0PSBVSbPCz2mzAxP3kZhMabkao/KuZu0IEI457aPEvTBhvRr98XYkB6cd6XIK/TUTyAIZQ4+DYmSP4+6W35d0="}',
        '2025-10-01 00:00:00', '2025-10-01 00:00:00'),
       ('01JAB9XED0GTXBSQDFVYAJ8WE5', '', NULL, 'oauth', 'google', 'Google', 'active', true, true,
        '{"org": "hivepaas-test", "clientId": "405460302846-8sk3j75rd5jonn9jfheis9coatkdn4jh.apps.googleusercontent.com", "clientSecret": "hpenc:PZPZhFUSouX/VYlc0SEK3Vcuaa/kO+K/N2k42PBStbBhvSjprJqfxO3EuDF212cRfu47Jvrb7gYjfmM9RX30"}',
        '2025-10-01 00:00:00', '2025-10-01 00:00:00'),
       ('01JAB9XED0GTXBSQDFVYAJ8WH1', '', NULL, 'im-service', 'slack', 'slack 1', 'active', true, true,
        '{"slack": {"webhook": "hpenc:VaHQxd7FplmGsTVwwPoUqeutAneORV2bjWCd0S9IzfPjeOYbSyAVVzzrPVJSHkjdCgXJ/Y14WquuwL7txbY2N0SuX7ATSg+z0tMzpeAIKQ1kWVTfn0K08QsO5LsM6vTF5ffwiwEucpv+f8Z6kw=="}}',
        '2025-10-01 00:00:00', '2025-10-01 00:00:00'),
       ('01JAB9XED0GTXBSQDFVYAJ8WI1', '', NULL, 'im-service', 'discord', 'discord 1', 'active', true, true,
        '{"discord": {"webhook": "hpenc:C2xBK2SKf9nEiyCSoVzPDKRDLYo1bG1Yss3TytQhSZdWTZEDPVPEgD1Ah3pKYVnbGvarIpaYVgIC+lej24JAaR+HpfFPgKGP90Y4gkFFIdfRO6hgCntmdgRt3qdNo1x0Z96zB5mLjCIR8saBzFOgbu4kvBsjuZ8ksBuYutHd5Q2FWUyuy9GWj33jE/G3V4BQsOAG9fE="}}',
        '2025-10-01 00:00:00', '2025-10-01 00:00:00'),
       ('01JAB9XED0GTXBSQDFVYAJ8WN1', '', NULL, 'email', 'smtp', 'email smtp 1', 'active', true, true,
        '{"smtp": {"ssl": false, "host": "smtp.gmail.com", "port": 587, "password": "hpenc:faWBZ6dY9Homx6ZWX6hgF2dN4yx2F+C8KF+twppWTSnzpJPefNlAPCDf+CE2pYg=", "username": "localpaas.dev@gmail.com", "displayName": "HivePaaS App"}}',
        '2025-10-01 00:00:00', '2025-10-01 00:00:00'),
       ('01JAB9XED0GTXBSQDFVYAJ8WP1', '', NULL, 'notification', NULL, 'notification 1', 'active', true, false,
        '{"viaEmail": {"enabled": true, "sender": {"id": "01JAB9XED0GTXBSQDFVYAJ8WN1"}, "toAllAdmins": true}, "viaSlack": {"enabled": true, "webhook": {"id": "01JAB9XED0GTXBSQDFVYAJ8WH1"}}, "viaDiscord": {"enabled": true, "webhook": {"id": "01JAB9XED0GTXBSQDFVYAJ8WI1"}}, "minSendInterval": "30s"}',
        '2025-10-01 00:00:00', '2025-10-01 00:00:00'),
       ('01JAB9XED0GTXBSQDFVYAJ8WP2', '', NULL, 'notification', NULL, 'notification tiendc@ only', 'active', true, true,
        '{"viaEmail": {"enabled": true, "sender": {"id": "01JAB9XED0GTXBSQDFVYAJ8WN1"}, "toAddresses": ["tiendc@gmail.com"]}, "viaSlack": {"enabled": true, "webhook": {"id": "01JAB9XED0GTXBSQDFVYAJ8WH1"}}, "viaDiscord": {"enabled": true, "webhook": {"id": "01JAB9XED0GTXBSQDFVYAJ8WI1"}}, "minSendInterval": "30s"}',
        '2025-10-01 00:00:00', '2025-10-01 00:00:00'),
       ('01JAB9XED0GTXBSQDFVYAJ8WU1', '', NULL, 'key-auth', '', 'r2 key', 'active', true, true,
        '{"keyId": "e74c319305880f8ce70dedb4463b4e8a", "secretKey": "hpenc:Wep2Rt6LyLSqEXFhx6WBpWCnKvSVftqlUpbReY24l/8716IEg7NXjVPTP0X8PIY0eaTjkDEQX405WbZGE2HJm4GmVo5Kj4OSI3VQ7RMQ17vDlj/yA3TmnScZJyQ="}',
        '2025-10-01 00:00:00', '2025-10-01 00:00:00'),
       ('01JAB9XED0GTXBSQDFVYAJ8WR1', '', NULL, 'cloud-storage', 'aws-s3', 'my cloud storage (s3)', 'active', true, true,
        '{"s3": {"keyAuth": {"id": "01JAB9XED0GTXBSQDFVYAJ8WU1"}, "bucket": "localpaas-dev", "region": "apac", "endpoint": "https://2d8acc928b0d56d7301b169067ba96af.r2.cloudflarestorage.com"}}',
        '2025-10-01 00:00:00', '2025-10-01 00:00:00')
ON CONFLICT DO NOTHING;

-- The organization a self-hosted server has, and its members
INSERT INTO organizations (id, slug, name, created_at, updated_at)
VALUES ('01JAB9XED0GTXBSQDFVYAJ8WO1', 'default', 'Default', '2025-10-01 00:00:00', '2025-10-01 00:00:00')
ON CONFLICT DO NOTHING;

INSERT INTO org_members (org_id, user_id, role, status, created_at, updated_at)
VALUES ('01JAB9XED0GTXBSQDFVYAJ8WO1', '01JAB9XED0GTXBSQDFVYAJ8WA1', 'owner', 'active', '2025-10-01 00:00:00', '2025-10-01 00:00:00'),
       ('01JAB9XED0GTXBSQDFVYAJ8WO1', '01JAB9XED0GTXBSQDFVYAJ8WA2', 'member', 'active', '2025-10-01 00:00:00', '2025-10-01 00:00:00'),
       ('01JAB9XED0GTXBSQDFVYAJ8WO1', '01JAB9XED0GTXBSQDFVYAJ8WA3', 'member', 'active', '2025-10-01 00:00:00', '2025-10-01 00:00:00')
ON CONFLICT DO NOTHING;
