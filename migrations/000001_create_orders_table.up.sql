CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    is_bid BOOLEAN NOT NULL,
    symbol VARCHAR NOT NULL,
    price VARCHAR NOT NULL,
    qty VARCHAR NOT NULL,
    type VARCHAR NOT NULL,
    size_filled VARCHAR NOT NULL DEFAULT '0',
    status VARCHAR NOT NULL DEFAULT 'filling',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    closed_at TIMESTAMP
);

CREATE TABLE matches (
    order_id INTEGER NOT NULL REFERENCES orders(id),
    order_id_counter INTEGER NOT NULL REFERENCES orders(id),
    qty VARCHAR NOT NULL,
    price VARCHAR NOT NULL,
    PRIMARY KEY (order_id, order_id_counter)
);

CREATE OR REPLACE FUNCTION update_order_status()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.qty = NEW.size_filled THEN
        NEW.status := 'filled';
        NEW.closed_at := CURRENT_TIMESTAMP;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_order_status
BEFORE UPDATE ON orders
FOR EACH ROW
EXECUTE FUNCTION update_order_status();