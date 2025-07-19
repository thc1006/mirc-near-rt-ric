/// <reference types="cypress" />

// Custom commands for xApp Dashboard testing

// Command to simulate login flow (even though this app doesn't have traditional auth)
Cypress.Commands.add('simulateLogin', () => {
  // Since there's no actual login, we just navigate to the front page
  // and then to home to simulate the user flow
  cy.visit('/frontpage');
  cy.get('app-front-page').should('be.visible');
  cy.visit('/home');
  cy.get('app-home').should('be.visible');
});

// Command to verify component loading with proper error handling
Cypress.Commands.add('verifyComponentLoads', (componentSelector: string, timeout = 10000) => {
  cy.get(componentSelector, { timeout }).should('be.visible');
  cy.get(componentSelector).should('not.have.class', 'error');
});

// Command to check table structure and data
Cypress.Commands.add('verifyTableStructure', (tableSelector: string, expectedHeaders: string[]) => {
  cy.get(tableSelector).should('be.visible');
  
  expectedHeaders.forEach((header, index) => {
    cy.get(`${tableSelector} th`).eq(index).should('contain.text', header);
  });
});

// Command to intercept and verify API calls
Cypress.Commands.add('interceptAndVerifyApi', (pattern: string, fixture?: string) => {
  const alias = `api_${Date.now()}`;
  
  if (fixture) {
    cy.intercept('GET', pattern, { fixture }).as(alias);
  } else {
    cy.intercept('GET', pattern).as(alias);
  }
  
  return cy.wrap(`@${alias}`);
});

// Command to check responsive design
Cypress.Commands.add('checkResponsiveDesign', () => {
  const viewports = [
    { width: 320, height: 568 },   // Mobile
    { width: 768, height: 1024 },  // Tablet
    { width: 1024, height: 768 },  // Desktop
    { width: 1920, height: 1080 }  // Large Desktop
  ];
  
  viewports.forEach((viewport) => {
    cy.viewport(viewport.width, viewport.height);
    cy.get('body').should('be.visible');
    cy.wait(500); // Allow time for responsive changes
  });
});

// Extend Cypress interface for TypeScript support
declare global {
  namespace Cypress {
    interface Chainable {
      /**
       * Simulate login flow for xApp Dashboard
       */
      simulateLogin(): Chainable<void>
      
      /**
       * Verify that a component loads properly
       */
      verifyComponentLoads(componentSelector: string, timeout?: number): Chainable<JQuery<HTMLElement>>
      
      /**
       * Verify table structure matches expected headers
       */
      verifyTableStructure(tableSelector: string, expectedHeaders: string[]): Chainable<void>
      
      /**
       * Intercept API calls and verify responses
       */
      interceptAndVerifyApi(pattern: string, fixture?: string): Chainable<string>
      
      /**
       * Check responsive design across multiple viewports
       */
      checkResponsiveDesign(): Chainable<void>
    }
  }
}