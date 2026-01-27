import type { Page } from '@playwright/test';
import { expect } from '@playwright/test';

export async function registerAndLoginAsAdmin(page: Page) {
    // 1. Go to /login - the proxy will handle redirects if setup is required
    await page.goto('/login');

    // 2. Check if we were redirected to /setup or stayed at /login
    if (page.url().includes('/setup')) {
        // --- REGISTRATION ---
        await expect(page.getByText('Welcome to JustBackup')).toBeVisible();

        await page.getByLabel('Username').fill('admin');
        await page.getByLabel('Password', { exact: true }).fill('password123');
        await page.getByLabel('Confirm Password').fill('password123');

        // Click Create Account
        await page.locator('button[type="submit"]').click();

        // After registration, we should be at login
        await page.waitForURL('**/login');
    } else {
        // Ensure we are at login
        await page.waitForURL('**/login');
    }

    // --- LOGIN ---
    await expect(page.getByText('Login to JustBackup')).toBeVisible();

    await page.getByLabel('Username').fill('admin');
    await page.getByLabel('Password', { exact: true }).fill('password123');

    const signInButton = page.locator('button[type="submit"]');
    await expect(signInButton).toBeVisible();
    await signInButton.click();

    // We check if there is an error with the credentials
    const errorMsg = page.locator('text=Invalid credentials');
    if (await errorMsg.isVisible({ timeout: 2000 })) {
        throw new Error('Login failed with "Invalid credentials"');
    }

    // Redirect to dashboard
    await page.waitForURL('**/dashboard');
}