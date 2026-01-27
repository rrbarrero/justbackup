import { test, expect } from '@playwright/test';
import { registerAndLoginAsAdmin } from './utils/auth';

test.describe.configure({ mode: 'serial' });

test.beforeEach(async ({ page }) => {
    await registerAndLoginAsAdmin(page);
});

async function createHost(page: any, hostName: string) {
    await page.getByRole('link', { name: 'Hosts' }).click();
    await page.getByRole('button', { name: 'Add Host' }).click();
    await page.getByLabel('Name', { exact: true }).fill(hostName);
    await page.getByLabel('Hostname / IP').fill('127.0.0.1');
    await page.getByLabel('SSH User').fill('root');
    await page.getByLabel('SSH Port').fill('22');
    await page.getByRole('button', { name: 'Add Host' }).click();
    await page.waitForURL('**/hosts');
}

test('should allow creating and editing backups with new cron syntax', async ({ page }) => {
    const timestamp = Date.now();
    const hostName = `Cron Test Host ${timestamp}`;

    // 1. Create a host
    await createHost(page, hostName);

    // Navigate to the new host
    const row = page.getByRole('row').filter({ hasText: hostName });
    await row.click();

    // 2. Create backup with @daily syntax
    await page.getByRole('button', { name: 'Add Backup' }).click();
    await page.getByLabel('Source Path').fill('/etc/cron-daily-test');
    await page.getByLabel('Source Path').blur();

    // Fill schedule with @daily
    await page.getByLabel('Schedule (Cron)').fill('@daily');

    // Create
    await page.getByRole('button', { name: 'Create Backup Task' }).click();

    // Verify it was created
    await expect(page.getByRole('heading', { name: 'Host Backups' })).toBeVisible();
    const backupRow = page.getByRole('row').filter({ hasText: '/etc/cron-daily-test' });
    await expect(backupRow).toBeVisible();

    // Verify schedule readability (optional, depends on UI)
    // For now we trust it was created.

    // 3. Edit backup to use ranges 1-5
    await backupRow.getByRole('button', { name: 'Edit' }).click();

    await expect(page.getByRole('heading', { name: 'Edit Backup Task' })).toBeVisible();

    // Check if @daily is in the input
    await expect(page.getByLabel('Schedule (Cron)')).toHaveValue('@daily');

    // Change to range
    await page.getByLabel('Schedule (Cron)').fill('0 0 * * 1-5'); // Mon-Fri
    await page.getByRole('button', { name: 'Save Changes' }).click();

    // Verify dialog closed
    await expect(page.getByRole('heading', { name: 'Edit Backup Task' })).not.toBeVisible();

    // Open details to confirm
    await backupRow.getByRole('cell', { name: '/etc/cron-daily-test' }).click();
    await expect(page.getByRole('heading', { name: 'Backup Details' })).toBeVisible();
    await expect(page.getByText('0 0 * * 1-5')).toBeVisible();
    await page.getByRole('button', { name: 'Close' }).click();
});
