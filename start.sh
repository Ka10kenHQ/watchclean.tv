#!/bin/bash
set -e

echo "Starting MongoDB..."
mongod --dbpath /data/db --bind_ip_all --fork --logpath /var/log/mongodb.log

echo "Waiting for MongoDB to be ready..."
until mongosh --eval "db.adminCommand('ping')" > /dev/null 2>&1; do
  echo "MongoDB is unavailable - sleeping"
  sleep 2
done

echo "MongoDB is up!"

# Check if database needs to be seeded
if mongosh mykadri --eval "db.movies.countDocuments()" --quiet | grep -q "^0$"; then
  echo "Restoring MongoDB data from seed files..."
  if [ -d "/mongo-seed/mykadri" ]; then
    mongorestore --db=mykadri /mongo-seed/mykadri/
    echo "MongoDB data restored successfully!"
  else
    echo "No seed data found, skipping restore"
  fi
else
  echo "Database already contains data, skipping seed"
fi

echo "Starting application..."
exec /app/scraper
