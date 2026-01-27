import { test, expect } from '@playwright/test';
import { registerAndLoginAsAdmin } from './utils/auth';

test.describe.configure({ mode: 'serial' });

test.beforeEach(async ({ page }) => {
    await registerAndLoginAsAdmin(page);
});

async function createHost(page: any, hostName: string) {
    // 1. Navigate to Hosts page
    await page.getByRole('link', { name: 'Hosts' }).click();
    await expect(page.getByRole('heading', { name: 'Hosts' })).toBeVisible();

    // 2. Click "Add Host"
    await page.getByRole('button', { name: 'Add Host' }).click();
    await expect(page.getByRole('heading', { name: 'New Host', exact: true })).toBeVisible();

    // 3. Fill out the form
    await page.getByLabel('Name', { exact: true }).fill(hostName);

    // Verify auto-fill of path
    const slug = hostName.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/(^-|-$)+/g, '');
    await expect(page.getByLabel('Host Path (Slug)')).toHaveValue(slug);

    await page.getByLabel('Hostname / IP').fill('ssh-source');
    await page.getByLabel('SSH User').fill('backup-test');
    await page.getByLabel('SSH Port').fill('22');

    // 4. Submit
    await page.getByRole('button', { name: 'Add Host' }).click();

    // 5. Verify redirection back to Hosts list
    await page.waitForURL('**/hosts');
    await expect(page.getByRole('heading', { name: 'Hosts' })).toBeVisible();
}

async function createBackup(page: any, hostName: string) {
    // Navigate to the host details page
    await page.getByRole('link', { name: 'Hosts' }).click();
    const row = page.getByRole('row').filter({ hasText: hostName });
    await row.click();

    // Click "Add Backup"
    await expect(page.getByRole('heading', { name: 'Host Backups' })).toBeVisible();
    await page.getByRole('button', { name: 'Add Backup' }).click();

    // Fill out the form
    await expect(page.getByRole('heading', { name: 'Create New Backup Task' })).toBeVisible();
    await page.getByLabel('Source Path').fill('/mnt/source_data');
    await page.getByLabel('Source Path').blur();
    await expect(page.getByLabel('Destination Name')).toHaveValue('source_data');
    await page.getByLabel('Destination Name').fill('website-backup');
    await page.getByLabel('Schedule (Cron)').fill('0 0 * * *');
    await page.getByLabel('Exclude Patterns').fill('node_modules/, .git/');

    // Submit
    await page.getByRole('button', { name: 'Create Backup Task' }).click();

    // Verify redirection back to Host details page
    await expect(page.getByRole('heading', { name: 'Host Backups' })).toBeVisible();
}

async function getBackupId(page: any): Promise<string> {
    let backupId = '';

    // Intercept the errors API call to extract backup ID
    page.on('request', (request: any) => {
        const url = request.url();
        const match = url.match(/\/api\/backups\/([^\/]+)\/errors/);
        if (match) {
            backupId = match[1];
        }
    });

    // Open and close the sheet to trigger the API call
    const backupRow = page.getByRole('row').filter({ hasText: '/mnt/source_data' });
    await backupRow.getByRole('cell').first().click();
    await page.waitForTimeout(500);
    await page.getByRole('button', { name: 'Close' }).click();
    await page.waitForTimeout(300);

    return backupId;
}

test('should not show Clear Logs button when there are no errors', async ({ page }) => {
    const timestamp = Date.now();
    const hostName = `No Errors Test ${timestamp}`;

    // 1. Setup: Create host and backup
    await createHost(page, hostName);
    await createBackup(page, hostName);

    // 2. Open backup details
    const backupRow = page.getByRole('row').filter({ hasText: '/mnt/source_data' });
    await backupRow.getByRole('cell', { name: '/mnt/source_data' }).click();

    // 3. Verify sheet opens
    await expect(page.getByRole('heading', { name: 'Backup Details' })).toBeVisible();
    const sheet = page.getByRole('dialog', { name: 'Backup Details' });

    // 4. Verify Error Logs section exists
    await expect(sheet.getByText('Error Logs', { exact: true })).toBeVisible();

    // 5. For a new backup without errors, Clear Logs button should NOT be visible
    const clearLogsButton = sheet.getByRole('button', { name: 'Clear Logs' });
    await expect(clearLogsButton).not.toBeVisible();

    // 6. Verify "No errors" message is shown
    await expect(sheet.getByText(/No errors found for this backup task/i)).toBeVisible();

    // 7. Close the sheet
    await page.getByRole('button', { name: 'Close' }).click();
    await expect(page.getByRole('heading', { name: 'Backup Details' })).not.toBeVisible();
});

test('should show Clear Logs button and confirmation dialog when errors exist', async ({ page, request }) => {
    const timestamp = Date.now();
    const hostName = `Clear Dialog Test ${timestamp}`;

    // 1. Setup: Create host and backup
    await createHost(page, hostName);
    await createBackup(page, hostName);

    // 2. Get backup ID
    const backupId = await getBackupId(page);

    if (!backupId) {
        throw new Error('Could not determine backup ID');
    }

    // 3. Seed test errors using request fixture
    await request.post('http://server:8080/api/v1/test/backup-errors/seed', {
        headers: {
            'Authorization': 'Basic YWRtaW46cGFzc3dvcmQxMjM=',
            'Content-Type': 'application/json',
        },
        data: [
            {
                backup_id: backupId,
                job_id: 'test-job-1',
                error_message: 'Connection timeout to remote server',
                occurred_at: new Date(Date.now() - 3600000).toISOString(),
            },
            {
                backup_id: backupId,
                job_id: 'test-job-2',
                error_message: 'Permission denied: /var/www/html/secret',
                occurred_at: new Date(Date.now() - 1800000).toISOString(),
            },
        ],
    });

    await page.waitForTimeout(500);

    // 4. NOW open backup details (errors should be loaded)
    const backupRow = page.getByRole('row').filter({ hasText: '/mnt/source_data' });
    await backupRow.getByRole('cell').first().click();
    await expect(page.getByRole('heading', { name: 'Backup Details' })).toBeVisible();

    const sheet = page.getByRole('dialog', { name: 'Backup Details' });

    // 5. Verify Clear Logs button is visible (errors were seeded)
    const clearLogsButton = sheet.getByRole('button', { name: 'Clear Logs' });
    await expect(clearLogsButton).toBeVisible({ timeout: 3000 });

    // 6. Verify error count badge
    await expect(sheet.getByText(/2 errors/)).toBeVisible();

    // 7. Click Clear Logs button
    await clearLogsButton.click();

    // 8. Verify confirmation dialog appears
    await expect(page.getByRole('heading', { name: 'Clear Error Logs?' })).toBeVisible();
    await expect(page.getByText(/This action cannot be undone/i)).toBeVisible();
    await expect(page.getByText(/permanently delete all error logs/i)).toBeVisible();

    // 9. Verify dialog has Cancel and Clear Logs buttons
    await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Clear Logs', exact: true })).toBeVisible();

    // 10. Click Cancel to close without deleting
    await page.getByRole('button', { name: 'Cancel' }).click();

    // 11. Verify confirmation dialog closes
    await expect(page.getByRole('heading', { name: 'Clear Error Logs?' })).not.toBeVisible();

    // 12. Verify Clear Logs button is still visible (errors not deleted)
    await expect(clearLogsButton).toBeVisible();

    // 13. Close the sheet
    await page.getByRole('button', { name: 'Close' }).click();
});

test('should successfully clear logs when confirmed', async ({ page, request }) => {
    const timestamp = Date.now();
    const hostName = `Clear Success Test ${timestamp}`;

    // 1. Setup: Create host and backup
    await createHost(page, hostName);
    await createBackup(page, hostName);

    // 2. Get backup ID
    const backupId = await getBackupId(page);

    if (!backupId) {
        throw new Error('Could not determine backup ID');
    }

    // 3. Seed test errors using request fixture
    await request.post('http://server:8080/api/v1/test/backup-errors/seed', {
        headers: {
            'Authorization': 'Basic YWRtaW46cGFzc3dvcmQxMjM=',
            'Content-Type': 'application/json',
        },
        data: [
            {
                backup_id: backupId,
                job_id: 'test-job-1',
                error_message: 'Test error for deletion',
                occurred_at: new Date().toISOString(),
            },
        ],
    });

    await page.waitForTimeout(500);


    // 4. Open backup details
    const backupRow = page.getByRole('row').filter({ hasText: '/mnt/source_data' });
    await backupRow.getByRole('cell').first().click();
    await expect(page.getByRole('heading', { name: 'Backup Details' })).toBeVisible();

    const sheet = page.getByRole('dialog', { name: 'Backup Details' });
    const clearLogsButton = sheet.getByRole('button', { name: 'Clear Logs' });

    // 5. Verify Clear Logs button is visible
    await expect(clearLogsButton).toBeVisible({ timeout: 3000 });

    // 6. Click Clear Logs
    await clearLogsButton.click();

    // 7. Confirm deletion
    await expect(page.getByRole('heading', { name: 'Clear Error Logs?' })).toBeVisible();
    const confirmButton = page.getByRole('button', { name: 'Clear Logs', exact: true });
    await confirmButton.click();

    // 8. Wait for deletion to complete and dialog to close
    await expect(page.getByRole('heading', { name: 'Clear Error Logs?' })).not.toBeVisible({ timeout: 5000 });

    // 9. Verify Clear Logs button is no longer visible
    await expect(clearLogsButton).not.toBeVisible();

    // 10. Verify success message appears
    await expect(sheet.getByText(/No errors found for this backup task/i)).toBeVisible();

    // 11. Close the sheet
    await page.getByRole('button', { name: 'Close' }).click();
});
