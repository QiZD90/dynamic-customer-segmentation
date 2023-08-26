INSERT INTO segments(slug) VALUES ('AVITO_VOICE_CHAT'); -- 1
INSERT INTO segments(slug) VALUES ('AVITO_SALE_30'); -- 2
INSERT INTO segments(slug) VALUES ('AVITO_SALE_50'); -- 3
INSERT INTO segments(slug) VALUES ('AVITO_PERFORMANCE_VAS'); -- 4

INSERT INTO users_segments(segment_id, user_id, expires_at) VALUES (1, 1000, NOW() + interval '3 minutes');
INSERT INTO users_segments(segment_id, user_id) VALUES (2, 1000);

INSERT INTO users_segments(segment_id, user_id) VALUES (1, 1004);

INSERT INTO users_segments(segment_id, user_id) VALUES (3, 1006);
INSERT INTO users_segments(segment_id, user_id) VALUES (4, 1006);

INSERT INTO users_segments(segment_id, user_id) VALUES (1, 1010);
INSERT INTO users_segments(segment_id, user_id) VALUES (2, 1010);
INSERT INTO users_segments(segment_id, user_id) VALUES (3, 1010);