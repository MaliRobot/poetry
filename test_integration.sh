#!/bin/bash

echo "Poetry Management System - Integration Test"
echo "==========================================="

# Test API server health
echo "Testing API server..."
curl -s http://localhost:8080/ping && echo " ✓ API server is healthy" || echo " ✗ API server is not responding"

# Test worker service health
echo "Testing worker service..."
curl -s http://localhost:8082/health && echo " ✓ Worker service is healthy" || echo " ✗ Worker service is not responding"

# Test worker status
echo "Checking worker status..."
curl -s http://localhost:8082/status | jq '.' && echo " ✓ Worker status retrieved" || echo " ✗ Could not get worker status"

# Test collections endpoint
echo "Testing collections endpoint..."
curl -s http://localhost:8080/collections | jq '.' && echo " ✓ Collections endpoint working" || echo " ✗ Collections endpoint failed"

# Test adding a single poem
echo "Testing single poem submission..."
curl -X POST http://localhost:8080/poem \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Integration Poem",
    "poem": "This is a test poem for integration testing",
    "poet": "Test Author",
    "language": "en",
    "dataset": "integration_test"
  }' && echo " ✓ Single poem submission working" || echo " ✗ Single poem submission failed"

# Create a test file for bulk upload
echo "Creating test bulk upload file..."
cat > /tmp/test_poems.json << EOF
[
  {
    "title": "Bulk Test Poem 1",
    "poem": "This is the first bulk test poem",
    "poet": "Bulk Test Author",
    "language": "en",
    "dataset": "bulk_test"
  },
  {
    "title": "Bulk Test Poem 2", 
    "poem": "This is the second bulk test poem",
    "poet": "Bulk Test Author",
    "language": "en",
    "dataset": "bulk_test"
  }
]
EOF

# Test bulk poem upload
echo "Testing bulk poem submission..."
curl -X POST http://localhost:8080/poems \
  -F "file=@/tmp/test_poems.json" && echo " ✓ Bulk poem submission working" || echo " ✗ Bulk poem submission failed"

# Check worker status after bulk upload
echo "Checking worker status after bulk upload..."
sleep 2
curl -s http://localhost:8082/status | jq '.'

echo "Integration test completed!"
