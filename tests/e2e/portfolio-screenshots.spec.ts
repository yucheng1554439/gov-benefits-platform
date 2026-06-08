import { test } from '@playwright/test';
import { ACCOUNTS, signIn } from './helpers/demo-ui';

const OUT = 'docs/screenshots';

test.describe('Portfolio screenshots', () => {
  test('capture key screens', async ({ page }) => {
    test.setTimeout(120_000);

    // Citizen Apply
    await signIn(page, ACCOUNTS.citizen);
    await page.goto('/citizen/apply');
    await page.getByRole('heading', { name: 'Apply for Benefits' }).waitFor();
    await page.screenshot({ path: `${OUT}/citizen-apply.png`, fullPage: true });

    // Worker Review (DEMO-A approved case with eligibility + benefit)
    await page.goto('/login');
    await signIn(page, ACCOUNTS.worker);
    await page.goto('/worker/queue');
    const demoCase = page.getByRole('link', { name: 'CASE-2026-DEMO-A' });
    if (await demoCase.count()) {
      await demoCase.click();
    } else {
      await page.getByRole('link', { name: /^CASE-2026-/ }).first().click();
    }
    await page.waitForURL(/\/worker\/cases\//);
    await page.screenshot({ path: `${OUT}/worker-review.png`, fullPage: true });

    // Benefit calculation close-up
    const benefitCard = page.getByRole('heading', { name: 'Benefit Amount' }).locator('xpath=ancestor::div[contains(@class,"rounded-lg")][1]');
    await benefitCard.screenshot({ path: `${OUT}/benefit-calculation.png` });

    // Supervisor Appeals
    await page.goto('/login');
    await signIn(page, ACCOUNTS.supervisor);
    await page.goto('/supervisor/appeals');
    await page.getByRole('heading', { name: 'Appeal Review' }).waitFor();
    await page.screenshot({ path: `${OUT}/appeals-review.png`, fullPage: true });

    // Admin Audit
    await page.goto('/login');
    await signIn(page, ACCOUNTS.admin);
    await page.goto('/admin/audit');
    await page.getByRole('heading', { name: 'Audit Trail' }).waitFor();
    await page.getByLabel('Filter by event type').selectOption('benefit.calculated');
    await page.waitForTimeout(800);
    await page.screenshot({ path: `${OUT}/audit-trail.png`, fullPage: true });

    // Analytics
    await page.goto('/dashboard/analytics');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: `${OUT}/analytics-dashboard.png`, fullPage: true });
  });
});
