#!/bin/bash

# This is a script that launches a deployment to test the load balancer.
# NOTE: Prod should use ansible instead of a bash script.

# Deploy Backends

docker-compose -f backend/docker-compose.yml up -d

poetry run python3 backend/build-backend-map.py

# Deploy Frontend

poetry run python3 frontend/server.py

# Deploy Load Balancer

go run ../lb/main.go testing/backend/backend-map.yml
