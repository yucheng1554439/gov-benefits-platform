import type { Locator, Page } from '@playwright/test';

/** Seed accounts from docs/demo-script.md */
export const ACCOUNTS = {
  citizen: 'citizen1@example.com',
  worker: 'worker1@dpss.lacounty.gov',
  supervisor: 'supervisor1@dpss.lacounty.gov',
  admin: 'admin@dpss.lacounty.gov',
} as const;

export const PASSWORD = 'Password123!';

/** Pause between major demo beats (ms). Override with DEMO_PAUSE_MS. */
export const PAUSE_MS = Number(process.env.DEMO_PAUSE_MS ?? 1800);

/** Extra delay on each action when not using Playwright slowMo. */
export const ACTION_MS = Number(process.env.DEMO_ACTION_MS ?? 350);

export async function pause(page: Page, ms: number = PAUSE_MS): Promise<void> {
  await page.waitForTimeout(ms);
}

export async function highlightClick(locator: Locator): Promise<void> {
  const page = locator.page();
  await locator.scrollIntoViewIfNeeded();
  await locator.evaluate((el) => {
    el.style.outline = '3px solid #fbbf24';
    el.style.outlineOffset = '3px';
    el.style.boxShadow = '0 0 0 8px rgba(251, 191, 36, 0.45)';
    el.style.transition = 'box-shadow 0.25s ease, outline 0.25s ease';
    el.style.position = 'relative';
    el.style.zIndex = '9999';
  });
  await page.waitForTimeout(ACTION_MS);
  await locator.click();
  await page.waitForTimeout(ACTION_MS);
}

export async function signIn(page: Page, email: string): Promise<void> {
  await page.goto('/login');
  await page.waitForLoadState('networkidle');
  await highlightClick(page.getByLabel('Email'));
  await page.getByLabel('Email').fill(email);
  await highlightClick(page.getByLabel('Password'));
  await page.getByLabel('Password').fill(PASSWORD);
  await highlightClick(page.getByRole('button', { name: 'Sign In' }));
  await page.waitForURL((url) => !url.pathname.startsWith('/login'), { timeout: 30_000 });
}

export async function signOut(page: Page): Promise<void> {
  await highlightClick(page.getByRole('button', { name: 'Sign Out' }));
  await page.waitForURL(/\/login/, { timeout: 15_000 });
}

export async function clickNav(page: Page, label: string): Promise<void> {
  await highlightClick(page.getByRole('link', { name: label, exact: true }));
  await page.waitForLoadState('networkidle');
}

export async function statusUpdate(page: Page, statusLabel: string): Promise<void> {
  const statusCard = page.getByRole('heading', { name: 'Update Status' }).locator('xpath=ancestor::div[contains(@class,"rounded-lg")][1]');
  const select = statusCard.locator('select');
  await select.scrollIntoViewIfNeeded();
  await select.selectOption({ label: statusLabel });
  await pause(page, 600);
  await highlightClick(statusCard.getByRole('button', { name: 'Update', exact: true }));
  await page.getByText(/Case status updated/i).waitFor({ timeout: 20_000 });
}

export async function openCaseInTable(page: Page, caseNumber: string): Promise<void> {
  await highlightClick(page.getByRole('link', { name: caseNumber }));
  await page.waitForURL(/\/cases\//, { timeout: 15_000 });
  await page.getByRole('heading', { name: caseNumber, level: 1 }).waitFor({ timeout: 15_000 });
}
