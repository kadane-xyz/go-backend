#!/bin/bash
set -e

# Judge0 Initialization

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE USER judge0 WITH SUPERUSER;
    ALTER USER "judge0" WITH PASSWORD '$JUDGE0_PASSWORD' SUPERUSER;
EOSQL

# Array of SQL files in order
sql_files=(
    "schema.sql"
    "accounts.sql"
    "friends.sql"
    "solutions.sql"
    "comments.sql"
    "problems.sql"
    "account_solved_problem.sql"
    "submissions.sql"
    "starred.sql"
)

# Loop through the files and execute them in order
for sql_file in "${sql_files[@]}"; do
    echo "Executing $sql_file..."
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -f "/docker-entrypoint-initdb.d/schemas/$sql_file"
done

# Insert sample data
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -f "/docker-entrypoint-initdb.d/data/sample.sql"
