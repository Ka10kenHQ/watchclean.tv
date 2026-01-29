#!/bin/bash
set -e

echo "Restoring MongoDB data from seed files..."

# Wait for MongoDB to be ready
until mongosh --eval "db.adminCommand('ping')" > /dev/null 2>&1; do
  echo "Waiting for MongoDB to start..."
  sleep 2
done

# Restore the database from BSON files
mongorestore --db=mykadri /docker-entrypoint-initdb.d/mykadri/

echo "MongoDB data restored successfully!"
