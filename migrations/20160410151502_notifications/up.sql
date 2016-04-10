
CREATE FUNCTION notify_event_trigger() RETURNS trigger AS $$
DECLARE
BEGIN
        PERFORM pg_notify('event_stream', row_to_json(NEW)::text);
        RETURN new;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER watched_events_trigger AFTER INSERT ON events
FOR EACH ROW EXECUTE PROCEDURE notify_event_trigger();
