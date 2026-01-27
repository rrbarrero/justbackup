import { test, expect } from '@playwright/test';
import { registerAndLoginAsAdmin } from './utils/auth';
import * as fs from 'fs';
import * as path from 'path';
import { execSync } from 'child_process';

test.describe.configure({ mode: 'serial' });

test.beforeEach(async ({ page }) => {
    await registerAndLoginAsAdmin(page);
});

async function createRemoteHost(page: any, name: string) {
    // 1. Navigate to Hosts page
    await page.getByRole('link', { name: 'Hosts' }).click();
    await expect(page.getByRole('heading', { name: 'Hosts' })).toBeVisible();

    // 2. Click "Add Host"
    await page.getByRole('button', { name: 'Add Host' }).click();
    await expect(page.getByRole('heading', { name: 'New Host', exact: true })).toBeVisible();

    // 3. Fill out the form
    await page.getByLabel('Name', { exact: true }).fill(name);
    await page.getByLabel('Hostname / IP').fill('ssh-source'); // Container name
    await page.getByLabel('SSH User').fill('backup-test');
    await page.getByLabel('SSH Port').fill('22');

    // 4. Submit
    await page.getByRole('button', { name: 'Add Host' }).click();

    // 5. Verify redirection back to Hosts list
    await page.waitForURL('**/hosts');
}

test('should execute a backup task and verify files on disk', async ({ page, request }) => {
    const timestamp = Date.now();
    const hostName = `Remote Source ${timestamp}`;

    // 1. Create Host
    await createRemoteHost(page, hostName);

    // 2. Navigate to Host details to create backup
    await page.getByRole('link', { name: 'Hosts' }).click();
    const hostRow = page.getByRole('row').filter({ hasText: hostName });
    await hostRow.click();

    // 3. Create Backup Task
    await expect(page.getByRole('heading', { name: 'Host Backups' })).toBeVisible();
    await page.getByRole('button', { name: 'Add Backup' }).click();

    await expect(page.getByRole('heading', { name: 'Create New Backup Task' })).toBeVisible();
    await page.getByLabel('Source Path').fill('/mnt/source_data'); // Path inside ssh-source container
    await page.getByLabel('Source Path').blur();

    // Override destination name
    const backupName = `e2e-test-${timestamp}`;
    await page.getByLabel('Destination Name').fill(backupName);
    await page.getByLabel('Schedule (Cron)').fill('0 0 * * *');

    // Submit
    await page.getByRole('button', { name: 'Create Backup Task' }).click();

    // 4. Verify backup appers in list
    await expect(page.getByRole('heading', { name: 'Host Backups' })).toBeVisible();
    const backupRow = page.getByRole('row').filter({ hasText: '/mnt/source_data' });
    await expect(backupRow).toBeVisible();

    // 5. Trigger Backup manually via UI
    const actionsCell = backupRow.getByRole('cell').last();
    // Use the first button in the actions cell which is usually "Run"
    await actionsCell.locator('button').first().click();

    // 6. Verify File System.
    const hostSlug = hostName.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/(^-|-$)+/g, '');
    const verifyPath = `/mnt/backups_verify/${hostSlug}/${backupName}`;

    // Polling for file existence
    await expect.poll(async () => {
        if (!fs.existsSync(verifyPath)) return false;
        const files = fs.readdirSync(verifyPath);
        return files.length > 0;
    }, {
        message: 'Backup files should be created on disk',
        timeout: 30000,
        intervals: [1000]
    }).toBe(true);

    // 7. Verify some content matches
    const files = fs.readdirSync(verifyPath);
    console.log('Backup files found:', files);

    let hasExampleFile = files.includes('example-file.txt');

    // If not in root, check inside source_data subdirectory (common rsync behavior)
    if (!hasExampleFile && files.includes('source_data')) {
        const subPath = path.join(verifyPath, 'source_data');
        const subFiles = fs.readdirSync(subPath);
        console.log('Backup subfiles found:', subFiles);
        hasExampleFile = subFiles.includes('example-file.txt');
    }

    expect(hasExampleFile).toBe(true);
});

test('should execute an encrypted backup task and verify decryption on disk', async ({ page, request }) => {
    const timestamp = Date.now();
    const hostName = `Crypto Source ${timestamp}`;

    // 1. Create Host
    await createRemoteHost(page, hostName);

    // 2. Navigate to Host details to create backup
    await page.getByRole('link', { name: 'Hosts' }).click();
    const hostRow = page.getByRole('row').filter({ hasText: hostName });
    await hostRow.click();

    // 3. Create Backup Task
    await expect(page.getByRole('heading', { name: 'Host Backups' })).toBeVisible();
    await page.getByRole('button', { name: 'Add Backup' }).click();

    await expect(page.getByRole('heading', { name: 'Create New Backup Task' })).toBeVisible();
    await page.getByLabel('Source Path').fill('/mnt/source_data');
    await page.getByLabel('Source Path').blur();

    const backupName = `enc-test-${timestamp}`;
    await page.getByLabel('Destination Name').fill(backupName);
    await page.getByLabel('Schedule (Cron)').fill('0 0 * * *');

    // Enable Encryption in Advanced Settings tab
    await page.getByRole('tab', { name: 'Advanced Settings' }).click();
    await page.getByLabel('Encrypt Backup').click();

    // Intercept creation request to get ID
    const createPromise = page.waitForResponse(response =>
        response.request().method() === 'POST' && response.url().includes('/backups'),
        { timeout: 60000 }
    );

    // Submit
    await page.getByRole('button', { name: 'Create Backup Task' }).click();

    const response = await createPromise;
    const body = await response.json();
    const backupId = body.id;
    expect(backupId).toBeDefined();

    // 4. Verify backup appears in list
    await expect(page.getByRole('heading', { name: 'Host Backups' })).toBeVisible();
    const backupRow = page.getByRole('row').filter({ hasText: '/mnt/source_data' });
    await expect(backupRow).toBeVisible();

    // 5. Trigger Backup manually via UI
    const actionsCell = backupRow.getByRole('cell').last();
    await actionsCell.locator('button').first().click();

    // 6. Verify File System (Encrypted file)
    const hostSlug = hostName.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/(^-|-$)+/g, '');
    const verifyPath = `/mnt/backups_verify/${hostSlug}/${backupName}.tar.gz.enc`;
    const unencryptedPath = `/mnt/backups_verify/${hostSlug}/${backupName}`;

    // Polling for file existence
    await expect.poll(async () => {
        return fs.existsSync(verifyPath) && !fs.existsSync(unencryptedPath);
    }, {
        message: 'Encrypted backup file should exist and unencrypted directory should be removed',
        timeout: 30000,
        intervals: [1000]
    }).toBe(true);

    console.log('Encrypted backup verified at:', verifyPath);

    // 7. Decrypt using CLI inside ssh-source
    const masterKey = process.env.ENCRYPTION_KEY;
    if (!masterKey) throw new Error('ENCRYPTION_KEY environment variable is not set');
    const encFilePath = `/mnt/backups/${hostSlug}/${backupName}.tar.gz.enc`;
    const decFilePath = `/mnt/decrypted/${backupName}.tar.gz`;

    // Command to run inside ssh-source
    const decryptCmd = `justbackup decrypt --file ${encFilePath} --out ${decFilePath} --id ${backupId} --key ${masterKey}`;

    // Preparing SSH key (must have 600 permissions)
    const privateKeyPath = '/tmp/backup_key';
    const tempKeyPath = '/tmp/id_ed25519_test';

    try {
        execSync(`cp ${privateKeyPath} ${tempKeyPath} && chmod 600 ${tempKeyPath}`);
        execSync(`ssh -i ${tempKeyPath} -o StrictHostKeyChecking=no backup-test@ssh-source "${decryptCmd}"`, { stdio: 'inherit' });
    } catch (error: any) {
        throw error;
    }

    // 8. Verify decrypted file exists
    const verifyDecryptedPath = `/mnt/decrypted/${backupName}.tar.gz`;
    expect(fs.existsSync(verifyDecryptedPath)).toBe(true);

    // Verify file is a valid tar.gz (at least check it's not empty)
    const stats = fs.statSync(verifyDecryptedPath);
    expect(stats.size).toBeGreaterThan(0);
    console.log('Decryption verified successfully at:', verifyDecryptedPath);
});

test('should execute an incremental backup task and verify multiple versions with hardlinks', async ({ page }) => {
    const timestamp = Date.now();
    const hostName = `Inc Host ${timestamp}`;
    const backupName = `inc-test-${timestamp}`;

    // 1. Setup Host
    await createRemoteHost(page, hostName);

    // 2. Navigate to Host details to create backup
    await page.getByRole('link', { name: 'Hosts' }).click();
    const hostRow = page.getByRole('row').filter({ hasText: hostName });
    await hostRow.click();

    // 3. Create Incremental Backup Task
    await expect(page.getByRole('heading', { name: 'Host Backups' })).toBeVisible();
    await page.getByRole('button', { name: 'Add Backup' }).click();

    await expect(page.getByRole('heading', { name: 'Create New Backup Task' })).toBeVisible();
    await page.getByLabel('Source Path').fill('/mnt/source_data');
    await page.getByLabel('Source Path').blur(); // Trigger autofill
    await page.waitForTimeout(1000); // Wait for potential race condition in autofill

    await page.getByLabel('Destination Name').fill(backupName);

    // Enable Incremental in Advanced Settings
    await page.getByRole('tab', { name: 'Advanced Settings' }).click();
    await page.getByLabel('Incremental Backup').click();

    await page.getByRole('button', { name: 'Create Backup Task' }).click();
    await expect(page.getByRole('heading', { name: 'Host Backups' })).toBeVisible();

    const backupRow = page.getByRole('row').filter({ hasText: '/mnt/source_data' });

    // 4. First Backup Run
    await backupRow.getByRole('cell').last().locator('button').first().click();

    // Wait for "completed" status in the table
    await expect(backupRow).toContainText('completed', { timeout: 30000 });

    // Verify disk structure for first run
    const hostSlug = hostName.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/(^-|-$)+/g, '');
    let backupBaseDir = `/mnt/backups_verify/${hostSlug}/${backupName}`;

    // Check for symlink with polling and fallback
    const latestSymlink = path.join(backupBaseDir, 'latest');

    await expect.poll(async () => {
        if (fs.existsSync(latestSymlink)) return true;

        // Fallback check for alternative paths if autofill interfered
        const possibleBase = `/mnt/backups_verify/${hostSlug}`;
        if (fs.existsSync(possibleBase)) {
            const dirs = fs.readdirSync(possibleBase);
            const found = dirs.find(d => d.includes(backupName));
            if (found) {
                backupBaseDir = path.join(possibleBase, found);
                return fs.existsSync(path.join(backupBaseDir, 'latest'));
            }
        }
        return false;
    }, {
        timeout: 15000,
        message: 'Latest symlink should be created in the backup directory'
    }).toBe(true);

    const firstVersion = fs.readlinkSync(path.join(backupBaseDir, 'latest'));
    const firstVersionPath = path.join(backupBaseDir, firstVersion);
    // Rsync without trailing slash creates a subdirectory with the name of the source dir
    const testFile = 'example-file.txt';

    // Verify file exists (check both paths/rsync behaviors)
    let firstFilePath = path.join(firstVersionPath, 'source_data', testFile);
    if (!fs.existsSync(firstFilePath)) {
        // Fallback: check if file is at root
        const rootFilePath = path.join(firstVersionPath, testFile);
        if (fs.existsSync(rootFilePath)) {
            firstFilePath = rootFilePath;
        } else {
            console.log(`File not found at ${firstFilePath} OR ${rootFilePath}. Content of ${firstVersionPath}:`);
            try { execSync(`ls -R ${firstVersionPath}`); } catch (e) { }
        }
    }
    expect(fs.existsSync(firstFilePath)).toBe(true);
    const firstInode = fs.statSync(firstFilePath).ino;

    // 5. Second Backup Run (No changes)
    await page.waitForTimeout(1200); // Ensure different timestamp
    await backupRow.getByRole('cell').last().locator('button').first().click();

    // Wait for symlink to point to a DIFFERENT directory
    await expect.poll(async () => {
        const currentSymlink = path.join(backupBaseDir, 'latest');
        if (!fs.existsSync(currentSymlink)) return false;
        return fs.readlinkSync(currentSymlink) !== firstVersion;
    }, { timeout: 30000, message: 'Symlink should update to a new version' }).toBe(true);

    const secondVersion = fs.readlinkSync(path.join(backupBaseDir, 'latest'));
    const secondVersionPath = path.join(backupBaseDir, secondVersion);
    const secondFilePath = path.join(secondVersionPath, 'source_data', testFile);
    expect(fs.existsSync(secondFilePath)).toBe(true);

    const secondInode = fs.statSync(secondFilePath).ino;

    // CRITICAL: Verify hardlink (same inode)
    expect(secondInode).toBe(firstInode);

    // 6. Add new file to source and run third backup
    const newFileName = `newfile-${timestamp}.txt`;
    const privateKeyPath = '/tmp/backup_key';
    const tempKeyPath = '/tmp/id_ed25519_inc';
    execSync(`cp ${privateKeyPath} ${tempKeyPath} && chmod 600 ${tempKeyPath}`);
    execSync(`ssh -i ${tempKeyPath} -o StrictHostKeyChecking=no backup-test@ssh-source "echo 'hello' > /mnt/source_data/${newFileName}"`);

    await page.waitForTimeout(1200); // Ensure different timestamp
    await backupRow.getByRole('cell').last().locator('button').first().click();

    await expect.poll(async () => {
        const currentSymlink = path.join(backupBaseDir, 'latest');
        if (!fs.existsSync(currentSymlink)) return false;
        return fs.readlinkSync(currentSymlink) !== secondVersion;
    }, { timeout: 30000, message: 'Symlink should update to a third version' }).toBe(true);

    const thirdVersion = fs.readlinkSync(path.join(backupBaseDir, 'latest'));
    const thirdVersionPath = path.join(backupBaseDir, thirdVersion);
    const thirdFilePath = path.join(thirdVersionPath, 'source_data', newFileName);
    expect(fs.existsSync(thirdFilePath)).toBe(true);
});
