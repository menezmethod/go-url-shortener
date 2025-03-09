#!/bin/sh
# Check if Swagger documentation is available

# Wait for the application to start
sleep 5

# Check the Swagger health endpoint
response=$(curl -s http://localhost:8081/api/swagger-health)
if echo "$response" | grep -q "Swagger documentation is available"; then
    echo "✅ Swagger documentation is available."
else
    echo "❌ Swagger documentation is not available!"
    echo "Response: $response"
    exit 1
fi

# Check if the doc.json file is accessible
doc_response=$(curl -s http://localhost:8081/swagger/doc.json)
if [ $? -eq 0 ] && [ -n "$doc_response" ]; then
    echo "✅ Swagger JSON file is accessible."
else
    echo "❌ Swagger JSON file is not accessible!"
    exit 1
fi

echo "✅ All Swagger checks passed!"
exit 0 