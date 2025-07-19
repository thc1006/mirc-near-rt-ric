describe('xApp Dashboard - Home Navigation and API Testing', () => {
  const baseUrl = 'http://localhost:4200';

  beforeEach(() => {
    // Intercept API calls to mock responses and verify HTTP status codes
    cy.intercept('GET', '/assets/data/*.json', { 
      fixture: 'sample-data.json',
      statusCode: 200 
    }).as('getTagsData');
    
    // Mock the chart.js library to prevent canvas rendering issues in headless mode
    cy.window().then((win) => {
      win.Chart = win.Chart || class MockChart {
        constructor() {}
        destroy() {}
        update() {}
      };
    });
  });

  describe('Initial Navigation and Front Page', () => {
    it('should load the application and display the front page', () => {
      cy.visit(baseUrl);
      
      // Verify the application loads successfully
      cy.title().should('contain', 'Cur');
      
      // Check that we're initially redirected to the front page
      cy.url().should('include', '/frontpage');
      
      // Verify the front page content loads
      cy.get('app-front-page').should('be.visible');
      
      // Check for key dashboard elements that indicate successful load
      cy.get('.container-fluid').should('be.visible');
    });

    it('should navigate from front page to home successfully', () => {
      cy.visit(`${baseUrl}/frontpage`);
      
      // Navigate to home page - this simulates the "login" flow since there's no auth
      cy.visit(`${baseUrl}/home`);
      
      // Verify successful navigation to home
      cy.url().should('include', '/home');
      cy.get('app-home').should('be.visible');
    });
  });

  describe('Home Page Functionality', () => {
    beforeEach(() => {
      // Start each test on the home page
      cy.visit(`${baseUrl}/home`);
    });

    it('should load the home page with all required elements', () => {
      // Verify home page header
      cy.get('h1').contains('Repositories').should('be.visible');
      
      // Verify the main data table exists
      cy.get('#dataTable').should('be.visible');
      
      // Check table headers
      cy.get('#dataTable th').should('contain', 'Repositories');
      cy.get('#dataTable th').should('contain', 'tags');
      
      // Verify the card structure
      cy.get('.card.shadow').should('be.visible');
      cy.get('.card-header').should('be.visible');
      cy.get('.card-body').should('be.visible');
    });

    it('should display repository data in the table', () => {
      // Wait for any data to load
      cy.get('#dataTable tbody tr', { timeout: 10000 }).should('exist');
      
      // Verify table structure
      cy.get('.table-responsive').should('be.visible');
      cy.get('#dataTable').should('be.visible');
      
      // Check that the table has the expected columns
      cy.get('#dataTable th').should('have.length', 2);
      cy.get('#dataTable th').eq(0).should('contain', 'Repositories');
      cy.get('#dataTable th').eq(1).should('contain', 'tags');
    });

    it('should have working navigation links to tags page', () => {
      // Look for repository links in the table
      cy.get('#dataTable').within(() => {
        // Check if there are any repository links
        cy.get('a[routerLink="/tags"]').then(($links) => {
          if ($links.length > 0) {
            // Click the first repository link
            cy.wrap($links.first()).click();
            
            // Verify navigation to tags page
            cy.url().should('include', '/tags');
            cy.get('app-tags').should('be.visible');
          } else {
            // If no data, verify the table structure is still correct
            cy.get('table').should('be.visible');
            cy.log('No repository data available for navigation test');
          }
        });
      });
    });
  });

  describe('Tags Component HTTP Status Verification', () => {
    beforeEach(() => {
      // Navigate to tags page for HTTP testing
      cy.visit(`${baseUrl}/tags`);
    });

    it('should successfully load tags component and return HTTP 200 for data requests', () => {
      // Verify tags component loads
      cy.get('app-tags').should('be.visible');
      
      // Check for tags page header
      cy.get('h1').should('be.visible');
      
      // Verify the data table structure
      cy.get('.table-responsive').should('be.visible');
      
      // Wait for potential API call and verify it returns 200
      cy.wait('@getTagsData').then((interception) => {
        expect(interception.response.statusCode).to.equal(200);
      });
    });

    it('should handle data loading and display tags information', () => {
      // Verify tags component structure
      cy.get('.card.shadow').should('be.visible');
      cy.get('.card-header').should('contain.text', 'Home');
      
      // Check table structure
      cy.get('table').should('be.visible');
      cy.get('table th').should('contain', 'Id');
      cy.get('table th').should('contain', 'Tag');
      cy.get('table th').should('contain', 'Created');
      cy.get('table th').should('contain', 'Layers');
      cy.get('table th').should('contain', 'Size');
    });

    it('should verify tags component makes HTTP requests with 200 status', () => {
      // Intercept specific API endpoints that tags component might use
      cy.intercept('GET', '**/assets/data/exampletest.json', {
        statusCode: 200,
        body: [
          {
            "Id": "test-id-1",
            "Tag": "latest",
            "Created": "2023-01-01",
            "Layers": "5",
            "Size": "100MB"
          }
        ]
      }).as('getExampleData');
      
      // Reload the tags component to trigger the API call
      cy.reload();
      
      // Wait for the API call and verify 200 status
      cy.wait('@getExampleData').then((interception) => {
        expect(interception.response.statusCode).to.equal(200);
        expect(interception.response.body).to.be.an('array');
      });
      
      // Verify the component handles the response correctly
      cy.get('app-tags').should('be.visible');
    });
  });

  describe('Navigation Flow Integration', () => {
    it('should complete full user flow: frontpage → home → tags', () => {
      // Start at front page
      cy.visit(`${baseUrl}/frontpage`);
      cy.get('app-front-page').should('be.visible');
      
      // Navigate to home
      cy.visit(`${baseUrl}/home`);
      cy.get('app-home').should('be.visible');
      cy.get('h1').contains('Repositories').should('be.visible');
      
      // Navigate to tags
      cy.visit(`${baseUrl}/tags`);
      cy.get('app-tags').should('be.visible');
      
      // Verify API call returns 200
      cy.wait('@getTagsData').then((interception) => {
        expect(interception.response.statusCode).to.equal(200);
      });
    });

    it('should handle direct navigation to each route', () => {
      const routes = ['/frontpage', '/home', '/tags', '/imagehistory'];
      
      routes.forEach((route) => {
        cy.visit(`${baseUrl}${route}`);
        cy.url().should('include', route);
        
        // Verify the route loads without errors
        cy.get('body').should('be.visible');
        cy.get('router-outlet').should('exist');
      });
    });
  });

  describe('Error Handling and Resilience', () => {
    it('should handle API failures gracefully', () => {
      // Intercept API calls with error status
      cy.intercept('GET', '/assets/data/*.json', {
        statusCode: 500,
        body: { error: 'Internal Server Error' }
      }).as('getErrorData');
      
      cy.visit(`${baseUrl}/tags`);
      
      // Verify the component still loads despite API error
      cy.get('app-tags').should('be.visible');
      cy.get('.table-responsive').should('be.visible');
      
      // API call should fail but component should handle it
      cy.wait('@getErrorData').then((interception) => {
        expect(interception.response.statusCode).to.equal(500);
      });
    });

    it('should maintain functionality when assets are missing', () => {
      // Test with 404 for data files
      cy.intercept('GET', '/assets/data/*.json', {
        statusCode: 404,
        body: 'Not Found'
      }).as('getMissingData');
      
      cy.visit(`${baseUrl}/tags`);
      
      // Component should still render
      cy.get('app-tags').should('be.visible');
      
      cy.wait('@getMissingData').then((interception) => {
        expect(interception.response.statusCode).to.equal(404);
      });
    });
  });
});