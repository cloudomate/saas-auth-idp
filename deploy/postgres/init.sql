-- Initialize databases for SaaS Starter Kit

-- Create Casdoor database
CREATE DATABASE casdoor;

-- Create OpenFGA database
CREATE DATABASE openfga;

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE casdoor TO saas;
GRANT ALL PRIVILEGES ON DATABASE openfga TO saas;
