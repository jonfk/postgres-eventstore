CREATE TABLE events (
  id BIGSERIAL PRIMARY KEY,
  event_id text NOT NULL,
  event_type text NOT NULL,
  event_offset bigint NOT NULL,
  timestamp timestamp without time zone NOT NULL,
  payload jsonb
);

CREATE INDEX idx_events_date_brin
       ON events
       USING BRIN (timestamp);


CREATE INDEX idx_events_id_brin
       ON events
       USING BRIN (id);
