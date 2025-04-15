CREATE SCHEMA IF NOT EXISTS quote;

SET search_path TO quote;

CREATE TABLE quote.candlestick (
    id SERIAL PRIMARY KEY,
    pair VARCHAR NOT NULL ,
    timeframe VARCHAR NOT NULL,
    openTime TIMESTAMP NOT NULL,
    closeTime TIMESTAMP NOT NULL,
    openPrice VARCHAR NOT NULL,
    closePrice VARCHAR NOT NULL,
    highPrice VARCHAR NOT NULL,
    lowPrice VARCHAR NOT NULL,
    turnover VARCHAR NOT NULL,
    volume VARCHAR NOT NULL
);

CREATE TABLE quote.pairs (
    pair VARCHAR PRIMARY KEY,
    orderBookPricePrecisions INTEGER[] NOT NULL,
    qtyPrecision INTEGER NOT NULL,
    candleStickTimeframes VARCHAR[] NOT NULL,
    createdAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO quote.pairs (pair, orderBookPricePrecisions, qtyPrecision, candleStickTimeframes)
VALUES
    ('BTC_USDT', ARRAY[2, 1, 0], 6, ARRAY['1s', '1m', '5m', '15m', '30m', '1h', '2h', '4h', '24h']);
    
CREATE SCHEMA IF NOT EXISTS matchingEngine;

SET search_path TO matchingEngine;

CREATE TABLE matchingEngine.orders (
    id SERIAL PRIMARY KEY,
    userID INTEGER NOT NULL,
    isBid BOOLEAN NOT NULL,
    pair VARCHAR NOT NULL REFERENCES quote.pairs(pair),
    price VARCHAR NOT NULL,
    qty VARCHAR NOT NULL,
    type VARCHAR NOT NULL,
    sizeFilled VARCHAR NOT NULL DEFAULT '0',
    status VARCHAR NOT NULL DEFAULT 'filling',
    createdAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    closedAt TIMESTAMP
);

CREATE TABLE matchingEngine.matches (
    orderID INTEGER NOT NULL REFERENCES matchingEngine.orders(id),
    orderIDCounter INTEGER NOT NULL REFERENCES matchingEngine.orders(id),
    qty VARCHAR NOT NULL,
    price VARCHAR NOT NULL,
    PRIMARY KEY (orderID, orderIDCounter)
);

CREATE OR REPLACE FUNCTION matchingEngine.update_order_status()
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
BEFORE UPDATE ON matchingEngine.orders
FOR EACH ROW
EXECUTE FUNCTION matchingEngine.update_order_status();