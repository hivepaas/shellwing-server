import type { ComponentType } from "react";

import { MODULE_IDS, ROUTE } from "@/application/shared/constants";
import { Navigate, Outlet, type RouteObject } from "react-router";

import { ModuleTitle } from "@application/shared/components/module-title";
import { ConditionalModule } from "@application/shared/permissions";

async function getLazyComponents() {
    return await import("./settings.module");
}

function createSettingsRoute(path: string, title: string, loadComponent: () => Promise<ComponentType>): RouteObject {
    return {
        path,
        element: (
            <ConditionalModule id={MODULE_IDS.Settings}>
                <ModuleTitle title={title}>
                    <Outlet />
                </ModuleTitle>
            </ConditionalModule>
        ),
        children: [
            {
                index: true,
                lazy: async () => {
                    const Component = await loadComponent();

                    return {
                        Component,
                    };
                },
            },
        ],
    };
}

function createSettingsModuleRoute(
    path: string,
    title: string,
    loadComponent: () => Promise<ComponentType>,
): RouteObject {
    return {
        path,
        element: (
            <ConditionalModule id={MODULE_IDS.Settings}>
                <ModuleTitle title={title}>
                    <Outlet />
                </ModuleTitle>
            </ConditionalModule>
        ),
        children: [
            {
                index: true,
                lazy: async () => {
                    const Component = await loadComponent();

                    return {
                        Component,
                    };
                },
            },
        ],
    };
}

export const settingsRouter: RouteObject = {
    lazy: async () => {
        const { SettingsDialogsContainer } = await getLazyComponents();

        return {
            element: (
                <>
                    <Outlet />
                    <SettingsDialogsContainer />
                </>
            ),
        };
    },
    children: [
        {
            path: ROUTE.settings.$pattern,
            element: (
                <ConditionalModule id={MODULE_IDS.Settings}>
                    <Navigate
                        to={ROUTE.settings.oauth.$route}
                        replace
                    />
                </ConditionalModule>
            ),
        },
        createSettingsRoute(ROUTE.settings.emailAccounts.$pattern, "Email Accounts", async () => {
            const { SettingsEmailAccountsRoute } = await getLazyComponents();

            return SettingsEmailAccountsRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.emailAccounts.create.$pattern, "Email Accounts", async () => {
            const { SettingsEmailAccountCreateRoute } = await getLazyComponents();

            return SettingsEmailAccountCreateRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.emailAccounts.edit.$pattern, "Email Accounts", async () => {
            const { SettingsEmailAccountEditRoute } = await getLazyComponents();

            return SettingsEmailAccountEditRoute;
        }),
        createSettingsRoute(ROUTE.settings.imPlatforms.$pattern, "IM Platforms", async () => {
            const { SettingsImPlatformsRoute } = await getLazyComponents();

            return SettingsImPlatformsRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.imPlatforms.create.$pattern, "IM Platforms", async () => {
            const { SettingsImPlatformCreateRoute } = await getLazyComponents();

            return SettingsImPlatformCreateRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.imPlatforms.edit.$pattern, "IM Platforms", async () => {
            const { SettingsImPlatformEditRoute } = await getLazyComponents();

            return SettingsImPlatformEditRoute;
        }),
        createSettingsRoute(ROUTE.settings.cloudStorages.$pattern, "Cloud Storages", async () => {
            const { SettingsCloudStoragesRoute } = await getLazyComponents();

            return SettingsCloudStoragesRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.cloudStorages.create.$pattern, "Cloud Storages", async () => {
            const { SettingsCloudStorageCreateRoute } = await getLazyComponents();

            return SettingsCloudStorageCreateRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.cloudStorages.edit.$pattern, "Cloud Storages", async () => {
            const { SettingsCloudStorageEditRoute } = await getLazyComponents();

            return SettingsCloudStorageEditRoute;
        }),
        createSettingsRoute(ROUTE.settings.oauth.$pattern, "OAuth", async () => {
            const { SettingsOAuthRoute } = await getLazyComponents();

            return SettingsOAuthRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.oauth.create.$pattern, "OAuth", async () => {
            const { SettingsOAuthCreateRoute } = await getLazyComponents();

            return SettingsOAuthCreateRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.oauth.edit.$pattern, "OAuth", async () => {
            const { SettingsOAuthEditRoute } = await getLazyComponents();

            return SettingsOAuthEditRoute;
        }),
        createSettingsRoute(ROUTE.settings.keyAuth.$pattern, "Key Auth", async () => {
            const { SettingsKeyAuthRoute } = await getLazyComponents();

            return SettingsKeyAuthRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.keyAuth.create.$pattern, "Key Auth", async () => {
            const { SettingsKeyAuthCreateRoute } = await getLazyComponents();

            return SettingsKeyAuthCreateRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.keyAuth.edit.$pattern, "Key Auth", async () => {
            const { SettingsKeyAuthEditRoute } = await getLazyComponents();

            return SettingsKeyAuthEditRoute;
        }),
        createSettingsRoute(ROUTE.settings.notificationTargets.$pattern, "Notification Targets", async () => {
            const { SettingsNotificationTargetsRoute } = await getLazyComponents();

            return SettingsNotificationTargetsRoute;
        }),
        createSettingsModuleRoute(
            ROUTE.settings.notificationTargets.create.$pattern,
            "Notification Targets",
            async () => {
                const { SettingsNotificationTargetCreateRoute } = await getLazyComponents();

                return SettingsNotificationTargetCreateRoute;
            },
        ),
        createSettingsModuleRoute(
            ROUTE.settings.notificationTargets.edit.$pattern,
            "Notification Targets",
            async () => {
                const { SettingsNotificationTargetEditRoute } = await getLazyComponents();

                return SettingsNotificationTargetEditRoute;
            },
        ),
    ],
} as const;
