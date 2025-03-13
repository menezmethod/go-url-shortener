# URL Shortener API - Postman Test Suite

This directory contains a comprehensive Postman collection for testing the URL Shortener API. The test suite includes pre-request scripts, post-request scripts, dynamic variables, and detailed assertions to validate API responses.

## Directory Structure

```
postman/
├── collections/                  # Postman collection files
│   ├── URL_Shortener_API_Auth.json               # Authentication endpoints
│   ├── URL_Shortener_API_Links_Part1.json        # Link creation and retrieval
│   ├── URL_Shortener_API_Links_Part2.json        # Link update and deletion
│   ├── URL_Shortener_API_Links_Part3.json        # Link statistics
│   └── URL_Shortener_API_Master.json             # Combined collection for workflow testing
└── environments/                 # Postman environment files
    └── URL_Shortener_API_Environment.json        # Environment variables
```

## Setup Instructions

1. Install [Postman](https://www.postman.com/downloads/) if you haven't already.
2. Import the collections and environment files:
   - Open Postman
   - Click on "Import" in the top left
   - Select "Folder" and navigate to the `postman/` directory
   - Alternatively, import each file individually

3. Configure the environment:
   - Select the imported "URL Shortener API Environment" from the environment dropdown in the top right
   - Click on the eye icon to view and edit environment variables
   - Update the following variables:
     - `baseUrl`: The base URL of your API (default: `http://localhost:8081`)
     - `masterPassword`: Your API's master password for authentication

## Using the Collections

### Master Collection

The `URL_Shortener_API_Master.json` collection provides a streamlined workflow for testing the entire API. The requests are organized in the order they should be executed:

1. **Authentication**: Get a JWT token for API access
2. **Link Management**: Create, retrieve, update, and delete short links

### Individual Collections

For more detailed testing or custom workflows, use the individual collections:

- **Auth**: Test authentication endpoints and JWT token generation
- **Links Part 1**: Test short link creation and retrieval operations
- **Links Part 2**: Test short link update and deletion operations
- **Links Part 3**: Test link statistics and analytics

## Key Features

### Dynamic Variables

The collection utilizes Postman environment variables to maintain state between requests:

- `authToken`: Stores the JWT token from the authentication request
- `shortCode`: Stores the code of the most recently created short link
- `customAlias`: Generates random aliases for link creation
- `expirationDate`: Sets future expiration dates for links

### Pre-request Scripts

Pre-request scripts are used to:

- Generate random test data
- Set up required variables
- Verify authentication status
- Ensure test dependencies are met

### Test Scripts

Test scripts validate API responses including:

- Status codes
- Response structure and data types
- Content validation
- JWT token verification
- Data consistency between requests
- Performance metrics (response time)

### Visualizations

The collections include visualization scripts for the statistics endpoint that render HTML representations of the link analytics data.

## Running Automated Tests

To run the entire test suite as an automated collection run:

1. Open the "URL_Shortener_API_Master" collection
2. Click the "Run" button in the collection overview
3. Ensure all requests are selected and the correct environment is chosen
4. Configure run settings (iterations, delays, etc.)
5. Click "Run" to execute the tests

## CI/CD Integration

These collections can be integrated with CI/CD pipelines using [Newman](https://learning.postman.com/docs/running-collections/using-newman-cli/command-line-integration-with-newman/), Postman's command-line collection runner:

```bash
# Install Newman
npm install -g newman

# Run the master collection with the environment
newman run ./postman/collections/URL_Shortener_API_Master.json -e ./postman/environments/URL_Shortener_API_Environment.json
```

## Debugging Failed Tests

If tests fail:

1. Check the Postman console (View > Show Postman Console) for detailed logs
2. Verify environment variables are set correctly
3. Ensure the API server is running and accessible
4. Confirm the master password is correct for authentication
5. Check if any previously created resources are still available

## Contributing

When adding new tests or modifying existing ones:

1. Use descriptive test names that indicate what's being tested
2. Include assertions that verify both positive and negative cases
3. Use console.log statements for debugging
4. Update environment variables as needed
5. Document any changes in this README 