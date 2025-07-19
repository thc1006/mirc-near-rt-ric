# xApp Dashboard E2E Tests

This directory contains Cypress end-to-end tests for the xApp Dashboard component of the O-RAN Near-RT RIC platform.

## Test Overview

The test suite (`cypress/e2e/home.spec.ts`) covers:

### 🔐 Authentication Flow
- **Navigation-based "login"**: Since the app doesn't have traditional authentication, tests simulate user flow from front page to home
- **Route verification**: Ensures proper routing between components

### 🏠 Home Page Verification
- **Component loading**: Verifies Home component renders correctly
- **Data table structure**: Validates repository table with proper headers
- **Navigation links**: Tests navigation from Home to Tags component
- **Dashboard elements**: Confirms all UI elements load properly

### 🏷️ Tags Component HTTP 200 Tests
- **API interception**: Mocks HTTP requests to `/assets/data/*.json`
- **Status code verification**: Ensures all API calls return HTTP 200
- **Error handling**: Tests graceful degradation on API failures
- **Data loading**: Verifies component handles JSON responses correctly

## Running the Tests

### Prerequisites
```bash
# Install dependencies (if not already done)
npm install

# Start the Angular development server
npm start
```

### Execute Tests
```bash
# Run tests in headless mode
npm run e2e

# Open Cypress Test Runner for interactive testing
npm run e2e:open

# Run tests for CI/CD pipeline
npm run e2e:ci

# Run both unit and e2e tests
npm run test:all
```

## Test Structure

```
cypress/
├── e2e/
│   └── home.spec.ts           # Main test specification
├── fixtures/
│   └── sample-data.json       # Mock data for API responses
├── support/
│   ├── commands.ts            # Custom Cypress commands
│   └── e2e.ts                 # Global configuration and setup
└── README.md                  # This file
```

## Key Test Features

### API Mocking
Tests use `cy.intercept()` to mock API responses:
```typescript
cy.intercept('GET', '/assets/data/*.json', { 
  fixture: 'sample-data.json',
  statusCode: 200 
}).as('getTagsData');
```

### Custom Commands
- `cy.simulateLogin()`: Navigates through the app's user flow
- `cy.verifyComponentLoads()`: Ensures components render without errors
- `cy.verifyTableStructure()`: Validates table headers and structure
- `cy.interceptAndVerifyApi()`: Simplifies API testing

### Error Handling
Tests verify the application gracefully handles:
- HTTP 404 (missing data files)
- HTTP 500 (server errors)
- Network timeouts
- Missing chart.js dependencies

## Test Scenarios

1. **Initial Navigation**: Front page → Home page flow
2. **Home Page Functionality**: Repository table and navigation
3. **Tags Component**: HTTP requests and data handling
4. **Full User Flow**: Complete navigation path testing
5. **Error Resilience**: Failure mode testing

## Configuration

### Cypress Config (`cypress.config.ts`)
- Base URL: `http://localhost:4200`
- Viewport: 1280x720
- Timeouts: 10 seconds
- Screenshots on failure: Enabled
- Video recording: Disabled (for performance)

### Environment Variables
- `apiTimeout`: 10000ms
- `retryCount`: 2

## CI/CD Integration

For automated testing in pipelines:
```bash
# Start application in background
npm start &

# Wait for server to be ready
npx wait-on http://localhost:4200

# Run tests
npm run e2e:ci
```

## Troubleshooting

### Common Issues
1. **Chart.js errors**: Tests include mocks for canvas rendering issues
2. **Angular not ready**: Custom `waitForAngular()` command handles timing
3. **API timeouts**: Configurable timeout values in cypress.config.ts

### Debug Mode
Use `npm run e2e:open` to run tests interactively and debug issues.

## O-RAN Context

These tests ensure the xApp Dashboard maintains the reliability required for O-RAN Near-RT RIC operations:
- **Sub-second response times**: Validates UI responsiveness
- **API reliability**: Ensures HTTP 200 responses for data integrity
- **Navigation flow**: Confirms user can access xApp management features
- **Error resilience**: Verifies graceful handling of network issues in telecommunications environments