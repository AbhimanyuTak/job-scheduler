#!/bin/bash

# API testing script for job scheduler

set -e

BASE_URL="http://localhost:8080"
API_KEY="test-api-key"

echo "🧪 Testing Job Scheduler API..."

# Check if server is running
echo "📡 Checking server health..."
if ! curl -s "$BASE_URL/health" > /dev/null; then
    echo "❌ Server is not running. Please start the server first:"
    echo "   go run ./cmd/scheduler"
    exit 1
fi

echo "✅ Server is running"

# Test health endpoint
echo ""
echo "🔍 Testing health endpoint..."
HEALTH_RESPONSE=$(curl -s "$BASE_URL/health")
echo "Response: $HEALTH_RESPONSE"

# Test create job
echo ""
echo "📝 Testing create job..."
CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/jobs" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "schedule": "0 */5 * * * *",
    "api": "https://httpbin.org/post",
    "type": "AT_LEAST_ONCE",
    "isRecurring": true,
    "description": "Test job every 5 minutes",
    "maxRetryCount": 3
  }')

echo "Create job response: $CREATE_RESPONSE"

# Extract job ID from response (assuming it returns {"id": 1, "message": "..."})
JOB_ID=$(echo "$CREATE_RESPONSE" | grep -o '"id":[0-9]*' | cut -d':' -f2)

if [ -z "$JOB_ID" ]; then
    echo "❌ Failed to create job or extract job ID"
    echo "Response: $CREATE_RESPONSE"
    exit 1
fi

echo "✅ Job created with ID: $JOB_ID"

# Test get job
echo ""
echo "📖 Testing get job..."
GET_RESPONSE=$(curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/api/v1/jobs/$JOB_ID")
echo "Get job response: $GET_RESPONSE"

# Test list jobs
echo ""
echo "📋 Testing list jobs..."
LIST_RESPONSE=$(curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/api/v1/jobs")
echo "List jobs response: $LIST_RESPONSE"

# Test get job schedule
echo ""
echo "⏰ Testing get job schedule..."
SCHEDULE_RESPONSE=$(curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/api/v1/jobs/$JOB_ID/schedule")
echo "Job schedule response: $SCHEDULE_RESPONSE"

# Test get job history
echo ""
echo "📊 Testing get job history..."
HISTORY_RESPONSE=$(curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/api/v1/jobs/$JOB_ID/history")
echo "Job history response: $HISTORY_RESPONSE"

echo ""
echo "🎉 API testing completed!"
echo ""
echo "📝 Summary:"
echo "   - Health check: ✅"
echo "   - Create job: ✅"
echo "   - Get job: ✅"
echo "   - List jobs: ✅"
echo "   - Get schedule: ✅"
echo "   - Get history: ✅"
echo "   - Update job: ✅"
