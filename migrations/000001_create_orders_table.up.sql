CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    userID INTEGER NOT NULL,
    isBid BOOLEAN NOT NULL,
    symbol VARCHAR NOT NULL,
    price VARCHAR NOT NULL,
    qty VARCHAR NOT NULL,
    type VARCHAR NOT NULL,
    sizeFilled VARCHAR NOT NULL DEFAULT '0',
    status VARCHAR NOT NULL DEFAULT 'filling',
    createdAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    closedAt TIMESTAMP
);

CREATE TABLE matches (
    orderID INTEGER NOT NULL REFERENCES orders(id),
    orderIDCounter INTEGER NOT NULL REFERENCES orders(id),
    qty VARCHAR NOT NULL,
    price VARCHAR NOT NULL,
    PRIMARY KEY (orderID, orderIDCounter)
);

CREATE OR REPLACE FUNCTION update_order_status()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.sizeFilled <> OLD.sizeFilled AND NEW.qty = NEW.sizeFilled THEN
        NEW.status := 'filled';
        NEW.closedAt := CURRENT_TIMESTAMP;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_order_status
BEFORE UPDATE ON orders
FOR EACH ROW
EXECUTE FUNCTION update_order_status();