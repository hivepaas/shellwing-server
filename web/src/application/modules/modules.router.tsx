import { Outlet, type RouteObject } from "react-router";

import { AppNavigate } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { CommonDialogsContainer } from "@application/shared/dialogs-container";
import { ModuleLayout } from "@application/shared/layouts/module";
import { Page404NotFound } from "@application/shared/pages";

import { operationsRouter } from "./operations";
import { settingsRouter } from "./settings/settings.router";
import { userManagementRouter } from "./user-management/user-management.router";

export const modulesRouter: RouteObject = {
    errorElement: (
        <ModuleLayout>
            <Page404NotFound />
        </ModuleLayout>
    ),
    element: (
        <ModuleLayout>
            <Outlet />
            <CommonDialogsContainer />
        </ModuleLayout>
    ),
    children: [
        // Where a signed-in user lands: their own profile, which everyone has.
        {
            path: ROUTE.home.$pattern,
            element: (
                <AppNavigate.Basic
                    to={ROUTE.currentUser.profile.$route}
                    replace
                    ignorePrevPath
                />
            ),
        },
        userManagementRouter,
        settingsRouter,
        operationsRouter,
    ],
} as const;
