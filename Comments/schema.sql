DROP TABLE IF EXISTS comments;

CREATE TABLE comments(
  id SERIAL PRIMARY KEY,
  post_id BIGINT NOT NULL,
  parent_id BIGINT DEFAULT 0,
  content TEXT NOT NULL,
  add_time BIGINT NOT NULL
);