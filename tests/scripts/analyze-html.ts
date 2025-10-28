#!/usr/bin/env bun
/**
 * HTML Analyzer - Extracts all interactive elements from HTML templates
 *
 * Purpose: Parse HTML files and identify all elements that need data-testid attributes
 *
 * Outputs:
 * - JSON map of element type → selector → metadata
 * - Used by validator and test helper generator
 */

import { parse } from 'node-html-parser';
import * as fs from 'fs';
import * as path from 'path';

interface InteractiveElement {
  type: 'button' | 'input' | 'link' | 'status' | 'message';
  testId: string | null;
  id: string | null;
  text: string | null;
  selector: string;
  line?: number;
}

interface AnalysisResult {
  file: string;
  totalElements: number;
  withTestId: number;
  withoutTestId: number;
  elements: InteractiveElement[];
  missingTestIds: InteractiveElement[];
}

/**
 * Analyze HTML file for interactive elements
 */
function analyzeHTML(filePath: string): AnalysisResult {
  const content = fs.readFileSync(filePath, 'utf-8');
  const root = parse(content);

  const elements: InteractiveElement[] = [];

  // Find all buttons
  root.querySelectorAll('button').forEach(btn => {
    elements.push({
      type: 'button',
      testId: btn.getAttribute('data-testid'),
      id: btn.getAttribute('id'),
      text: btn.textContent?.trim() || null,
      selector: 'button',
    });
  });

  // Find all links (that act as buttons)
  root.querySelectorAll('a.gcp-btn, a[onclick]').forEach(link => {
    elements.push({
      type: 'link',
      testId: link.getAttribute('data-testid'),
      id: link.getAttribute('id'),
      text: link.textContent?.trim() || null,
      selector: 'a',
    });
  });

  // Find all inputs
  root.querySelectorAll('input[type="text"], input[type="password"], input[type="email"]').forEach(input => {
    elements.push({
      type: 'input',
      testId: input.getAttribute('data-testid'),
      id: input.getAttribute('id'),
      text: input.getAttribute('placeholder'),
      selector: 'input',
    });
  });

  // Find all status badges
  root.querySelectorAll('.status-badge').forEach(badge => {
    elements.push({
      type: 'status',
      testId: badge.getAttribute('data-testid'),
      id: badge.getAttribute('id'),
      text: badge.textContent?.trim() || null,
      selector: '.status-badge',
    });
  });

  // Find error/success messages
  root.querySelectorAll('.error-message, .success-message').forEach(msg => {
    elements.push({
      type: 'message',
      testId: msg.getAttribute('data-testid'),
      id: msg.getAttribute('id'),
      text: null,
      selector: msg.classList.contains('error-message') ? '.error-message' : '.success-message',
    });
  });

  const withTestId = elements.filter(e => e.testId !== null);
  const withoutTestId = elements.filter(e => e.testId === null);

  return {
    file: path.basename(filePath),
    totalElements: elements.length,
    withTestId: withTestId.length,
    withoutTestId: withoutTestId.length,
    elements,
    missingTestIds: withoutTestId,
  };
}

/**
 * Main execution
 */
function main() {
  const htmlPath = path.join(__dirname, '../../pkg/server/templates/gcp_tool.html');

  if (!fs.existsSync(htmlPath)) {
    console.error(`❌ HTML file not found: ${htmlPath}`);
    process.exit(1);
  }

  console.log('🔍 Analyzing HTML for interactive elements...\n');

  const result = analyzeHTML(htmlPath);

  console.log(`📄 File: ${result.file}`);
  console.log(`📊 Total interactive elements: ${result.totalElements}`);
  console.log(`✅ With data-testid: ${result.withTestId}`);
  console.log(`❌ Missing data-testid: ${result.withoutTestId}\n`);

  if (result.missingTestIds.length > 0) {
    console.log('⚠️  Elements missing data-testid attributes:\n');
    result.missingTestIds.forEach(el => {
      console.log(`   ${el.type.toUpperCase()}: ${el.text || el.id || '(no identifier)'}`);
      if (el.id) console.log(`     → ID: ${el.id}`);
    });
    console.log('');
  }

  // Save results to JSON for other scripts
  const outputPath = path.join(__dirname, '../.cache/html-analysis.json');
  fs.mkdirSync(path.dirname(outputPath), { recursive: true });
  fs.writeFileSync(outputPath, JSON.stringify(result, null, 2));

  console.log(`💾 Analysis saved to: ${outputPath}\n`);

  // Return appropriate exit code
  if (result.missingTestIds.length > 0) {
    console.log('⚠️  Some elements are missing data-testid attributes');
    console.log('   Run the validator for detailed guidance.\n');
  } else {
    console.log('✅ All interactive elements have data-testid attributes!\n');
  }
}

main();
