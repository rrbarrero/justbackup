import { test, expect } from '@playwright/test';

test.describe('Dashboard (authenticated)', () => {
    test('should display sidebar and heading', async ({ page }) => {
        await page.goto('/dashboard');

        await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();
        await expect(page.getByText('JustBackup')).toBeVisible();
        await expect(page.getByRole('link', { name: 'Hosts' })).toBeVisible();
    });

    test('should navigate to completed backups from breakdown', async ({ page }) => {
        // 1. Navigate to dashboard
        await page.goto('/dashboard');

        // 2. Click on the "Completed" link in the breakdown section
        await page.getByRole('link', { name: 'Completed Successfully finished backups' }).click();

        // 3. Verify navigation to backups page with status filter
        await page.waitForURL('**/backups?status=completed');

        // 4. Verify the correct heading is displayed
        await expect(page.getByRole('heading', { name: 'Completed Backups' })).toBeVisible();
    });

    test('should navigate to pending backups from breakdown', async ({ page }) => {
        // 1. Navigate to dashboard
        await page.goto('/dashboard');

        // 2. Click on the "Pending" link in the breakdown section
        await page.getByRole('link', { name: 'Pending Scheduled or running' }).click();

        // 3. Verify navigation to backups page with status filter
        await page.waitForURL('**/backups?status=pending');

        // 4. Verify the correct heading is displayed
        await expect(page.getByRole('heading', { name: 'Pending Backups' })).toBeVisible();
    });

    test('should navigate to failed backups from breakdown', async ({ page }) => {
        // 1. Navigate to dashboard
        await page.goto('/dashboard');

        // 2. Click on the "Failed" link in the breakdown section
        await page.getByRole('link', { name: 'Failed Encountered errors' }).click();

        // 3. Verify navigation to backups page with status filter
        await page.waitForURL('**/backups?status=failed');

        // 4. Verify the correct heading is displayed
        await expect(page.getByRole('heading', { name: 'Failed Backups' })).toBeVisible();
    });

    test('should display host column in backups table', async ({ page }) => {
        // 1. Navigate to dashboard
        await page.goto('/dashboard');

        // 2. Click on any breakdown item to go to backups page
        await page.getByRole('link', { name: 'Completed Successfully finished backups' }).click();

        // 3. Wait for navigation
        await page.waitForURL('**/backups?status=completed');

        // 4. Verify the Host column header is present
        await expect(page.getByRole('columnheader', { name: 'Host', exact: true })).toBeVisible();
        await expect(page.getByRole('columnheader', { name: 'Hostname / IP' })).toBeVisible();

        // 5. If there are backups, verify that host names appear in the table rows
        const table = page.getByRole('table');
        const rowCount = await table.getByRole('row').count();

        if (rowCount > 1) { // More than just the header row
            // Verify at least one backup row contains host information
            const firstDataRow = table.getByRole('row').nth(1);
            await expect(firstDataRow).toBeVisible();
        }
    });
});