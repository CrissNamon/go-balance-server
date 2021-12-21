CREATE TABLE IF NOT EXISTS transactions (
	id SERIAL PRIMARY KEY,
	account INTEGER NOT NULL CHECK (account > 0),
	operation INTEGER NOT NULL DEFAULT 0,
	description VARCHAR(256) NOT NULL DEFAULT '',
	sum NUMERIC(16, 2) NOT NULL, 
	date BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM (now() AT TIME ZONE 'UTC'))
);

CREATE INDEX IF NOT EXISTS transaction_account ON transactions(account);
CREATE INDEX IF NOT EXISTS transaction_account_date ON transactions(account, date);
CREATE INDEX IF NOT EXISTS transactions_account_sum ON transactions(account, sum);
CREATE INDEX IF NOT EXISTS transactions_account_sum_date ON transactions(account, sum, date);

CREATE DATABASE "compose-postgres-test" WITH TEMPLATE "compose-postgres" OWNER "compose-postgres";