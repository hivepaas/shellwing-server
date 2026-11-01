import { MODULE_IDS, ROUTE } from "@/application/shared/constants";
import type { RouteObject } from "react-router";

import { AppNavigate } from "@application/shared/components";
import { ModuleTitle } from "@application/shared/components/module-title";
import { ConditionalModule } from "@application/shared/permissions";

async function getLazyComponents() {
    return await import("./operations.module");
}

export const operationsRouter: RouteObject = {
    children: [
        {
            path: ROUTE.operations.$pattern,
            element: (
                <AppNavigate.Basic
                    to={ROUTE.operations.tasks.$route}
                    replace
                    ignorePrevPath
                />
            ),
        },
        {
            path: ROUTE.operations.tasks.$pattern,
            lazy: async () => {
                const { SystemTasksRoute } = await getLazyComponents();

                return {
                    element: (
                        <ConditionalModule id={MODULE_IDS.System}>
                            <ModuleTitle title="Tasks">
                                <SystemTasksRoute />
                            </ModuleTitle>
                        </ConditionalModule>
                    ),
                };
            },
        },
        {
            path: ROUTE.operations.tasks.details.$pattern,
            lazy: async () => {
                const { SystemTaskDetailsRoute } = await getLazyComponents();

                return {
                    element: (
                        <ConditionalModule id={MODULE_IDS.System}>
                            <ModuleTitle title="Task Details">
                                <SystemTaskDetailsRoute />
                            </ModuleTitle>
                        </ConditionalModule>
                    ),
                };
            },
        },
        {
            path: ROUTE.operations.auditLogs.$pattern,
            lazy: async () => {
                const { SystemAuditLogsRoute } = await getLazyComponents();

                return {
                    element: (
                        <ConditionalModule id={MODULE_IDS.System}>
                            <ModuleTitle title="Audit Logs">
                                <SystemAuditLogsRoute />
                            </ModuleTitle>
                        </ConditionalModule>
                    ),
                };
            },
        },
    ],
} as const;
