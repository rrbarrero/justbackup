// e2e/global-setup.ts
import { chromium, type FullConfig } from '@playwright/test';
import { registerAndLoginAsAdmin } from './tests/utils/auth';

async function globalSetup(config: FullConfig) {
    // Coge el baseURL del primer proyecto (ajusta si usas otro)
    const baseURL = config.projects[0].use.baseURL as string | undefined;

    // IMPORTANTE: crea el contexto con baseURL
    const browser = await chromium.launch();
    const context = await browser.newContext({
        baseURL: baseURL ?? 'http://web:3000', // o lo que uses en Docker
    });
    const page = await context.newPage();

    await registerAndLoginAsAdmin(page);

    // Guarda el estado de la sesión
    await context.storageState({ path: 'playwright/.auth/admin.json' });

    await browser.close();
}
export default globalSetup;
