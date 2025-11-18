CREATE TABLE IF NOT EXISTS subscriptions (
    id SERIAL PRIMARY KEY,
    service_name TEXT NOT NULL,
    price INTEGER NOT NULL,
    user_id TEXT NOT NULL,
    start_date TEXT NOT NULL,
    end_date TEXT
);
