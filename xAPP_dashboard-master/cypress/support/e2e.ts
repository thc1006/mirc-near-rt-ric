// ***********************************************************
// This example support/e2e.ts is processed and
// loaded automatically before your test files.
//
// This is a great place to put global configuration and
// behavior that modifies Cypress.
//
// You can change the location of this file or turn off
// automatically serving support files with the
// 'supportFile' configuration option.
//
// You can read more here:
// https://on.cypress.io/configuration
// ***********************************************************

// Import commands.ts using ES2015 syntax:
import './commands'

// Alternatively you can use CommonJS syntax:
// require('./commands')

// Global configuration for xApp Dashboard E2E tests
Cypress.on('uncaught:exception', (err, runnable) => {
  // Prevent Cypress from failing on uncaught exceptions
  // This is useful for Angular applications that might have
  // harmless console errors during development
  if (err.message.includes('Chart is not defined') || 
      err.message.includes('ResizeObserver loop limit exceeded')) {
    return false;
  }
  // Return true to fail the test on other uncaught exceptions
  return true;
});

// Custom command to wait for Angular to be ready
declare global {
  namespace Cypress {
    interface Chainable {
      /**
       * Wait for Angular application to be ready
       */
      waitForAngular(): Chainable<void>
      
      /**
       * Navigate to a route and wait for it to load
       */
      navigateToRoute(route: string): Chainable<void>
      
      /**
       * Check HTTP status of an API call
       */
      checkApiStatus(url: string, expectedStatus: number): Chainable<void>
    }
  }
}

// Add custom commands for better test reliability
Cypress.Commands.add('waitForAngular', () => {
  cy.window().then((win: any) => {
    if (win.ng) {
      cy.window().its('ng').invoke('getTestability', win.document.body)
        .invoke('whenStable').then(() => {
          cy.log('Angular is stable');
        });
    } else {
      // Fallback for when Angular testing utilities aren't available
      cy.wait(100);
    }
  });
});

Cypress.Commands.add('navigateToRoute', (route: string) => {
  cy.visit(route);
  cy.waitForAngular();
  cy.url().should('include', route.replace(/^\//, ''));
});

Cypress.Commands.add('checkApiStatus', (url: string, expectedStatus: number) => {
  cy.request({
    url: url,
    failOnStatusCode: false
  }).then((response) => {
    expect(response.status).to.eq(expectedStatus);
  });
});

// Setup default intercepts for common API patterns
beforeEach(() => {
  // Mock common chart libraries to prevent canvas issues
  cy.window().then((win) => {
    if (!win.Chart) {
      win.Chart = class MockChart {
        constructor() {}
        destroy() {}
        update() {}
        resize() {}
      };
    }
  });
});