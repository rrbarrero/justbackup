import { test, expect } from '@playwright/test';
import { registerAndLoginAsAdmin } from './utils/auth';

test.describe.configure({ mode: 'serial' });

test('should allow user to register and login', async ({ page }) => {
    await registerAndLoginAsAdmin(page);

    // Check sidebar is visible
    await expect(page.getByText('JustBackup')).toBeVisible();
    await expect(page.getByRole('link', { name: 'Hosts' })).toBeVisible();
});

test('should show login form', async ({ page }) => {
    await page.goto('/login');

    // Check for the presence of the login form elements
    await expect(page.getByLabel('Username')).toBeVisible();
    await expect(page.getByLabel('Password')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Sign In' })).toBeVisible();
});
