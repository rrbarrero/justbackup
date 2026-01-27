import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
    testDir: './tests',
    globalSetup: './global-setup.ts',
    fullyParallel: true,
    forbidOnly: !!process.env.CI,
    retries: process.env.CI ? 2 : 0,
    workers: process.env.CI ? 1 : undefined,
    reporter: 'html',
    use: {
        baseURL: 'http://web:3000',
        trace: 'on-first-retry',
    },
    projects: [
        {
            name: 'unauthenticated',
            testMatch: /login\.spec\.ts/,
            use: {
                storageState: undefined, // without auth
                // ...devices['Desktop Chrome']
            },
        },
        // Proyecto "autenticado": usa el estado guardado
        {
            name: 'authenticated',
            // todos los demás .spec.ts excepto login.spec.ts
            testMatch: /^(?!.*login\.spec\.ts$).*\.spec\.ts$/,
            use: {
                storageState: 'playwright/.auth/admin.json',
            },
        },
    ],
});

