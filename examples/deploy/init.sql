-- Initialize databases for Examples

-- Create OpenFGA database
CREATE DATABASE openfga;
GRANT ALL PRIVILEGES ON DATABASE openfga TO examples;

-- Create Casdoor database
CREATE DATABASE casdoor;
GRANT ALL PRIVILEGES ON DATABASE casdoor TO examples;
