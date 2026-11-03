import type { RouteObject } from "react-router";

import { AppNavigate } from "@application/shared/components";
import { ModuleTitle } from "@application/shared/components/module-title";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { ConditionalModule } from "@application/shared/permissions";

async function getLazyComponents() {
    return await import("./system-settings.module");
}

export const systemSettingsRouter: RouteObject = {
    children: [
        {
            path: ROUTE.system.$pattern,
            element: (
                <AppNavigate.Basic
                    to={ROUTE.system.license.$route}
                    replace
                    ignorePrevPath
                />
            ),
        },
        {
            path: ROUTE.system.license.$pattern,
            lazy: async () => {
                const { SystemLicenseRoute } = await getLazyComponents();

                return {
                    element: (
                        <ConditionalModule id={MODULE_IDS.System}>
                            <ModuleTitle title="License">
                                <SystemLicenseRoute />
                            </ModuleTitle>
                        </ConditionalModule>
                    ),
                };
            },
        },
    ],
} as const;
