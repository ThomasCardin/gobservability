-- Initial database setup for gobservability
-- This file is automatically executed when PostgreSQL starts

-- Create extensions if needed
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Set timezone
SET timezone = 'UTC';

-- The tables will be auto-created by GORM migrations
-- This file can be used for any initial data or additional setup

-- Example: Create initial admin user or default alert rules
-- INSERT INTO alert_rules (id, node_name, target, metric, operator, threshold, duration_seconds, enabled, created_at, updated_at) 
-- VALUES (uuid_generate_v4(), 'default', 'node', 'cpu', '>', 80.0, 300, true, NOW(), NOW());