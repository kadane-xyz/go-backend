#!/bin/bash

files=(*.sql)
loaded=()
attempts=0
max_attempts=10

while [ ${#files[@]} -gt 0 ] && [ $attempts -lt $max_attempts ]; do
    failed=()
    
    files_shuffled=($(printf "%s\n" "${files[@]}" | shuf))
    
    echo -e "\nAttempt $((attempts + 1)) - Trying files in new order:"
    printf '%s\n' "${files_shuffled[@]}"
    
    for file in "${files_shuffled[@]}"; do
        if PGPASSWORD=kadane psql -h localhost -p 5432 -f "$file" -U kadane kadane ; then
            echo "✓ Successfully loaded $file"
            loaded+=("$file")
        else
            echo "✗ Failed to load $file - will retry in different order"
            failed+=("$file")
        fi
    done
    
    files=("${failed[@]}")
    ((attempts++))
done

if [ ${#files[@]} -eq 0 ]; then
    echo -e "\nAll schemas loaded successfully after $attempts attempts"
    exit 0
else
    echo -e "\nFailed to load all schemas after $attempts attempts."
    echo "Remaining problematic files:"
    printf '%s\n' "${files[@]}"
    exit 1
fi