import { test, expect } from '@playwright/test';
import {
  ACCOUNTS,
  clickNav,
  highlightClick,
  openCaseInTable,
  pause,
  signIn,
  signOut,
  statusUpdate,
} from './helpers/demo-ui';

/**
 * Option A — Live Walkthrough from docs/demo-script.md
 *
 * Uses reset_demo_data (via `npm run demo`) for a clean database.
 * Sections 5–6 apply only when the supervisor denies; this recording
 * follows the approve + letter path (sections 1–4 and 7).
 */
test.describe('Option A — Live Walkthrough', () => {
  test('recorded demo flow', async ({ page }) => {
    test.setTimeout(300_000);

    let caseNumber = '';
    let caseId = '';

    // --- 1. Citizen applies ---
    await test.step('1. Citizen applies', async () => {
      await signIn(page, ACCOUNTS.citizen);
      await pause(page);

      await clickNav(page, 'Apply');
      await expect(page.getByRole('heading', { name: 'Apply for Benefits' })).toBeVisible();

      // Step 1: Agency + Program (exact labels — header also has an agency picker)
      await page.getByLabel('Agency', { exact: true }).selectOption({ label: 'LA County DPSS' });
      await page.getByLabel('Program', { exact: true }).selectOption({ label: 'Food Assistance' });
      await highlightClick(page.getByRole('button', { name: 'Continue' }));

      // Step 2: Household
      await page.getByLabel('Household Size').fill('3');
      await highlightClick(page.getByRole('button', { name: 'Continue' }));

      // Step 3: Income
      await page.getByLabel('Annual Income ($)').fill('24000');
      await highlightClick(page.getByRole('button', { name: 'Continue' }));

      // Step 4: Location (ZIP required by the wizard)
      await page.getByLabel('ZIP Code').fill('90001');
      await highlightClick(page.getByRole('button', { name: 'Continue' }));

      // Step 5: Review + Submit
      await expect(page.getByText('Food Assistance')).toBeVisible();
      await highlightClick(page.getByRole('button', { name: 'Submit Application' }));
      await page.waitForURL(/\/citizen\/cases\/([^/]+)/, { timeout: 30_000 });
      caseId = page.url().match(/\/citizen\/cases\/([^/]+)/)?.[1] ?? '';
      expect(caseId).toBeTruthy();

      // Open from dashboard per demo script
      await clickNav(page, 'Dashboard');
      caseNumber = (await page.getByRole('link', { name: /^CASE-2026-/ }).first().textContent())?.trim() ?? '';
      expect(caseNumber).toMatch(/^CASE-2026-/);
      await openCaseInTable(page, caseNumber);
      await pause(page);
    });

    // --- 2. Worker reviews ---
    await test.step('2. Worker reviews', async () => {
      await signOut(page);
      await signIn(page, ACCOUNTS.worker);
      await pause(page);

      await clickNav(page, 'Queue');
      await openCaseInTable(page, caseNumber);

      const statusCard = page.getByRole('heading', { name: 'Update Status' }).locator('xpath=ancestor::div[contains(@class,"rounded-lg")][1]');
      await highlightClick(statusCard.getByRole('button', { name: 'Fraud Scan' }));
      await page.getByText(/Fraud scan complete/i).waitFor({ timeout: 20_000 });
      await pause(page);

      const eligibilityCard = page.getByRole('heading', { name: 'Eligibility' }).locator('xpath=ancestor::div[contains(@class,"rounded-lg")][1]');
      await highlightClick(eligibilityCard.getByRole('button', { name: 'Evaluate Eligibility' }));
      await eligibilityCard.getByText('Applicant is eligible.').waitFor({ timeout: 20_000 });
      await pause(page);

      const benefitCard = page.getByRole('heading', { name: 'Benefit Amount' }).locator('xpath=ancestor::div[contains(@class,"rounded-lg")][1]');
      await highlightClick(benefitCard.getByRole('button', { name: 'Calculate Benefit' }));
      await page.getByText(/\$\d+/).first().waitFor({ timeout: 20_000 });
      await pause(page);

      await statusUpdate(page, 'Under Review');
      await pause(page);
      await statusUpdate(page, 'Eligibility Review');
      await pause(page);
      // Required so the case appears under Supervisor → Escalations with Approve/Deny actions
      await statusUpdate(page, 'Supervisor Review');
      await pause(page);
    });

    // --- 3. Supervisor approves ---
    await test.step('3. Supervisor approves', async () => {
      await signOut(page);
      await signIn(page, ACCOUNTS.supervisor);
      await pause(page);

      await clickNav(page, 'Escalations');
      await expect(page.getByRole('heading', { name: 'Escalations' })).toBeVisible();
      await openCaseInTable(page, caseNumber);

      await highlightClick(page.getByRole('button', { name: 'Approve' }));
      await page.getByText(/Case status updated to Approved/i).waitFor({ timeout: 20_000 });
      await pause(page);
    });

    // --- 4. Worker generates letter ---
    await test.step('4. Worker generates letter', async () => {
      await signOut(page);
      await signIn(page, ACCOUNTS.worker);
      await pause(page);

      // Approved cases are not listed in the worker queue; open the case directly.
      await page.goto(`/worker/cases/${caseId}`);
      await page.getByRole('heading', { name: caseNumber, level: 1 }).waitFor({ timeout: 15_000 });

      const correspondence = page.getByRole('heading', { name: 'Correspondence' }).locator('xpath=ancestor::div[contains(@class,"rounded-lg")][1]');
      await highlightClick(correspondence.getByRole('button', { name: 'Generate Approval Letter' }));
      await page.getByText('Letter generated successfully.').waitFor({ timeout: 20_000 });
      await pause(page);

      await signOut(page);
      await signIn(page, ACCOUNTS.citizen);
      await pause(page);

      await clickNav(page, 'Dashboard');
      await openCaseInTable(page, caseNumber);
      await highlightClick(page.getByRole('link', { name: 'View Letters' }));
      await page.getByRole('heading', { name: 'Case Letters' }).waitFor();

      const downloadPromise = page.waitForEvent('download');
      await highlightClick(page.getByRole('button', { name: 'Download PDF' }));
      const download = await downloadPromise;
      expect(download.suggestedFilename()).toMatch(/letter-.*\.pdf/i);
      await pause(page);
    });

    // --- 7. Admin audit ---
    await test.step('7. Admin audit', async () => {
      await signOut(page);
      await signIn(page, ACCOUNTS.admin);
      await pause(page);

      await clickNav(page, 'Audit Trail');
      await expect(page.getByRole('heading', { name: 'Audit Trail' })).toBeVisible();

      await highlightClick(page.getByLabel('Filter by event type'));
      await page.getByLabel('Filter by event type').selectOption('benefit.calculated');
      await pause(page, 1200);
      await expect(page.getByText('benefit calculated').first()).toBeVisible({ timeout: 20_000 });

      await page.getByLabel('Filter by event type').selectOption('letter.generated');
      await pause(page, 1200);
      await expect(page.getByText('letter generated').first()).toBeVisible({ timeout: 20_000 });
      await pause(page);
    });
  });
});
