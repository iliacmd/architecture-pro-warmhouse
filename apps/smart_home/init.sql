-- Create the database if it doesn't exist
CREATE DATABASE smarthome;

-- Connect to the database
\c smarthome;

-- Create the sensors table
CREATE TABLE IF NOT EXISTS sensors (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL,
    location VARCHAR(100) NOT NULL,
    value FLOAT DEFAULT 0,
    unit VARCHAR(20),
    status VARCHAR(20) NOT NULL DEFAULT 'inactive',
    last_updated TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_sensors_type ON sensors(type);
CREATE INDEX IF NOT EXISTS idx_sensors_location ON sensors(location);
CREATE INDEX IF NOT EXISTS idx_sensors_status ON sensors(status);

-- Create the database if it doesn't exist
CREATE DATABASE telemetry;

-- Connect to the database
\c telemetry;

CREATE EXTENSION IF NOT EXISTS timescaledb;

CREATE TABLE telemetry_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE, -- 'temperature'
    description TEXT,
    data_type VARCHAR(20) NOT NULL CHECK (data_type IN ('number', 'boolean', 'string', 'json')),
    unit VARCHAR(50), -- 'celsius'
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE telemetry_data (
    time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    device_id SERIAL NOT NULL,
    telemetry_type_id SERIAL NOT NULL REFERENCES telemetry_types(id),
    numeric_value DOUBLE PRECISION,
    boolean_value BOOLEAN,
    string_value TEXT,
    json_value JSONB,
    PRIMARY KEY (time, device_id, telemetry_type_id)
);

SELECT create_hypertable(
    'telemetry_data',
    'time',
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- Индексы для telemetry_data
CREATE INDEX idx_telemetry_data_device_time
ON telemetry_data (device_id, time DESC);

CREATE INDEX idx_telemetry_data_type_time
ON telemetry_data (telemetry_type_id, time DESC);

CREATE INDEX idx_telemetry_data_device_type_time
ON telemetry_data (device_id, telemetry_type_id, time DESC);

CREATE INDEX idx_telemetry_data_time_desc
ON telemetry_data (time DESC);

SELECT add_retention_policy('telemetry_data', INTERVAL '1 year');

INSERT INTO telemetry_types (name, description, data_type, unit) VALUES
    ('temperature', 'Temperature reading', 'number', 'celsius')
    ('device_status', 'Device status', 'string', NULL)

INSERT INTO telemetry_data (time,device_id,telemetry_type_id,numeric_value) VALUES(now(),1,1,22.2)

-- Create the database if it doesn't exist
CREATE DATABASE device;

-- Connect to the database
\c device;

CREATE TABLE device_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE, -- 'Термостат'
    capabilities JSONB,
    protocol_type VARCHAR(50) -- 'HTTP'
);

CREATE TYPE status_enum AS ENUM ('inactive', 'active');

CREATE TYPE connection_status_enum AS ENUM ('connected', 'disconnected');

CREATE TABLE IF NOT EXISTS devices (
    id SERIAL PRIMARY KEY,
    device_type_id SERIAL NOT NULL REFERENCES device_types(id),
    name VARCHAR(100) NOT NULL UNIQUE, -- "Датчик температуры в гостиной"
    location VARCHAR(100) NOT NULL, -- "LivingRoom"
    status status_enum DEFAULT 'active',
    connection_status connection_status_enum DEFAULT 'connected',
    connection_info JSONB,-- (протокол, endpoint, credentials)
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
