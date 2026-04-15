CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL UNIQUE,
    transaction_id UUID NOT NULL,
    amount BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL
    );