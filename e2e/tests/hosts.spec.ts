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

test('should allow user to create a new host', async ({ page }) => {
    const timestamp = Date.now();
    const hostName = `Test Server ${timestamp}`;

    await createHost(page, hostName);

    // 6. Verify the new host is visible in the table
    const row = page.getByRole('row').filter({ hasText: hostName });
    await expect(row).toBeVisible();
    await expect(row).toContainText('ssh-source');
    await expect(row).toContainText('backup-test');
    await expect(row).toContainText('22');
});

async function createBackup(page: any, hostName: string) {
    // Navigate to the host details page
    // We assume we are on the hosts list or can get there
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

test('should allow user to add a backup task to a host', async ({ page }) => {
    const timestamp = Date.now();
    const hostName = `Backup Host ${timestamp}`;

    // 1. Create a host first
    await createHost(page, hostName);

    // 2. Create backup using the helper (which also verifies the flow)
    await createBackup(page, hostName);

    // 3. Verify the new backup task is visible in the list
    // Note: destination and schedule columns are hidden by default
    await expect(page.getByText('/mnt/source_data')).toBeVisible();
});

test('should allow user to edit a backup task', async ({ page }) => {
    const timestamp = Date.now();
    const hostName = `Edit Backup Host ${timestamp}`;

    // 1. Setup: Create host and backup
    await createHost(page, hostName);
    await createBackup(page, hostName);

    // 2. Find the backup row and click Edit
    const backupRow = page.getByRole('row').filter({ hasText: '/mnt/source_data' });
    await backupRow.getByRole('button', { name: 'Edit' }).click();

    // 3. Verify Dialog and Update Schedule
    await expect(page.getByRole('heading', { name: 'Edit Backup Task' })).toBeVisible();
    await page.getByLabel('Schedule (Cron)').fill('0 12 * * *');
    await page.getByRole('button', { name: 'Save Changes' }).click();

    // 4. Verify Dialog closes and schedule can be verified in details sheet
    await expect(page.getByRole('heading', { name: 'Edit Backup Task' })).not.toBeVisible();

    // Open details sheet to verify schedule was updated
    await backupRow.getByRole('cell', { name: '/mnt/source_data' }).click();
    await expect(page.getByRole('heading', { name: 'Backup Details' })).toBeVisible();
    await expect(page.getByText('At 12:00 PM')).toBeVisible();
    await page.getByRole('button', { name: 'Close' }).click();
});

test('should allow user to delete a backup task', async ({ page }) => {
    const timestamp = Date.now();
    const hostName = `Delete Backup Host ${timestamp}`;

    // 1. Setup: Create host and backup
    await createHost(page, hostName);
    await createBackup(page, hostName);

    // 2. Find the backup row and click Delete
    const backupRow = page.getByRole('row').filter({ hasText: '/mnt/source_data' });
    await backupRow.getByRole('button', { name: 'Delete' }).click();

    // 3. Verify Dialog and Confirm
    await expect(page.getByRole('heading', { name: 'Are you sure?' })).toBeVisible();
    await page.getByRole('button', { name: 'Delete' }).click();

    // 4. Verify Dialog closes and backup is removed
    await expect(page.getByRole('heading', { name: 'Are you sure?' })).not.toBeVisible();
    await expect(backupRow).not.toBeVisible();
});

test('should allow user to run a backup task immediately', async ({ page }) => {
    const timestamp = Date.now();
    const hostName = `Run Backup Host ${timestamp}`;

    // 1. Setup: Create host and backup
    await createHost(page, hostName);
    await createBackup(page, hostName);

    // 2. Find the backup row
    const backupRow = page.getByRole('row').filter({ hasText: '/mnt/source_data' });

    // 3. Click "Run" button
    await backupRow.getByRole('button', { name: 'Run' }).click();

    // 4. Verify status changes to "completed"
    // Note: The worker simulates success with a 1s delay in dev mode.
    // We wait for the status badge to show "completed"
    await expect(backupRow).toContainText('completed', { timeout: 10000 });

    // 5. Verify "Last Backup" is updated (not "Never")
    await expect(backupRow).not.toContainText('Never');
});

test('should allow user to delete a host', async ({ page }) => {
    const timestamp = Date.now();
    const hostName = `Delete Host ${timestamp}`;

    // 1. Setup: Create host
    await createHost(page, hostName);

    // 2. Find the host row and click Delete
    const row = page.getByRole('row').filter({ hasText: hostName });
    await row.getByRole('button', { name: 'Delete' }).click();

    // 3. Verify Dialog and Confirm
    await expect(page.getByRole('heading', { name: 'Are you sure?' })).toBeVisible();
    await expect(page.getByText(`This will permanently delete the host ${hostName}`)).toBeVisible();
    await page.getByRole('button', { name: 'Delete' }).click();

    // 4. Verify Dialog closes and host is removed
    await expect(page.getByRole('heading', { name: 'Are you sure?' })).not.toBeVisible();
    await expect(row).not.toBeVisible();
});

test('should display backup details in a sheet when clicking on a backup row', async ({ page }) => {
    const timestamp = Date.now();
    const hostName = `Details Host ${timestamp}`;

    // 1. Setup: Create host and backup
    await createHost(page, hostName);
    await createBackup(page, hostName);

    // 2. Find the backup row (excluding the actions column)
    const backupRow = page.getByRole('row').filter({ hasText: '/mnt/source_data' });

    // 3. Click on the row (not on a button) to open the details sheet
    // Click on the path cell specifically to avoid clicking on action buttons
    // We target the cell itself to ensure the click is registered on a non-interactive element that bubbles up
    await backupRow.getByRole('cell', { name: '/mnt/source_data' }).click();

    // 4. Verify the sheet opens with the correct heading
    await expect(page.getByRole('heading', { name: 'Backup Details' })).toBeVisible();

    // 5. Verify backup information is displayed
    const sheet = page.getByRole('dialog', { name: 'Backup Details' });
    await expect(sheet.getByText('Backup Configuration')).toBeVisible();
    await expect(sheet.getByText('/mnt/source_data')).toBeVisible();
    await expect(sheet.getByText('website-backup')).toBeVisible();

    // 6. Verify schedule information
    await expect(sheet.getByText('Schedule')).toBeVisible();
    await expect(sheet.getByText('0 0 * * *')).toBeVisible();
    await expect(sheet.getByText('At 12:00 AM')).toBeVisible(); // cronstrue translation

    // 7. Verify error logs section is present
    await expect(sheet.getByText('Error Logs', { exact: true })).toBeVisible();

    // 8. Verify no errors message (since this backup hasn't failed)
    await expect(sheet.getByText(/No errors found for this backup task/i)).toBeVisible();

    // 9. Close the sheet by clicking the X button
    await page.getByRole('button', { name: 'Close' }).click();

    // 10. Verify sheet is closed
    await expect(page.getByRole('heading', { name: 'Backup Details' })).not.toBeVisible();
});

test('should not open details sheet when clicking on action buttons', async ({ page }) => {
    const timestamp = Date.now();
    const hostName = `Actions Test Host ${timestamp}`;

    // 1. Setup: Create host and backup
    await createHost(page, hostName);
    await createBackup(page, hostName);

    // 2. Find the backup row
    const backupRow = page.getByRole('row').filter({ hasText: '/mnt/source_data' });

    // 3. Click the Edit button
    await backupRow.getByRole('button', { name: 'Edit' }).click();

    // 4. Verify Edit dialog opens, not the details sheet
    await expect(page.getByRole('heading', { name: 'Edit Backup Task' })).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Backup Details' })).not.toBeVisible();

    // 5. Close the edit dialog
    await page.getByRole('button', { name: 'Cancel' }).click();

    // 6. Verify edit dialog closes
    await expect(page.getByRole('heading', { name: 'Edit Backup Task' })).not.toBeVisible();
});

test('should display exclude patterns in backup details sheet', async ({ page }) => {
    const timestamp = Date.now();
    const hostName = `Excludes Test Host ${timestamp}`;

    // 1. Setup: Create host and backup with excludes
    await createHost(page, hostName);
    await createBackup(page, hostName);

    // 2. Click on the backup row to open details
    const backupRow = page.getByRole('row').filter({ hasText: '/mnt/source_data' });
    await backupRow.getByText('/mnt/source_data').click();

    // 3. Verify sheet opens
    await expect(page.getByRole('heading', { name: 'Backup Details' })).toBeVisible();

    // 4. Verify exclude patterns section shows count
    await expect(page.getByText('2 patterns')).toBeVisible();

    // 5. Verify individual exclude patterns are displayed
    await expect(page.getByText('node_modules/')).toBeVisible();
    await expect(page.getByText('.git/')).toBeVisible();

    // 6. Close the sheet
    await page.getByRole('button', { name: 'Close' }).click();
});
