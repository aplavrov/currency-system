CREATE DATABASE "wallet-db";
CREATE DATABASE "exchange-db";

CREATE USER "wallet-user" WITH ENCRYPTED PASSWORD 'wallet-password';
CREATE USER "exchange-user" WITH ENCRYPTED PASSWORD 'exchange-password';

ALTER USER "wallet-user" WITH SUPERUSER;
ALTER USER "exchange-user" WITH SUPERUSER;