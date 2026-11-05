import { Outlet, type RouteObject } from "react-router";

import { AppNavigate } from "@application/shared/components";
import { ModuleTitle } from "@application/shared/components/module-title";
import { ROUTE } from "@application/shared/constants";

async function getLazyComponents() {
    return await import("./organization.module");
}

/**
 * What an organization's owners and admins manage. Not behind a module permission: who may is the organization's
 * role, which the route itself checks.
 */
export const organizationRouter: RouteObject = {
    lazy: async () => {
        const { OrganizationDialogsContainer } = await getLazyComponents();

        return {
            element: (
                <>
                    <Outlet />
                    <OrganizationDialogsContainer />
                </>
            ),
        };
    },
    children: [
        {
            path: ROUTE.organization.$pattern,
            element: (
                <AppNavigate.Basic
                    to={ROUTE.organization.members.$route}
                    replace
                    ignorePrevPath
                />
            ),
        },
        {
            path: ROUTE.organization.members.$pattern,
            lazy: async () => {
                const { MembersRoute } = await getLazyComponents();

                return {
                    element: (
                        <ModuleTitle title="Members">
                            <MembersRoute />
                        </ModuleTitle>
                    ),
                };
            },
        },
        {
            path: ROUTE.organization.groups.$pattern,
            lazy: async () => {
                const { GroupsRoute } = await getLazyComponents();

                return {
                    element: (
                        <ModuleTitle title="Groups">
                            <GroupsRoute />
                        </ModuleTitle>
                    ),
                };
            },
        },
        {
            path: ROUTE.organization.sshCertificates.$pattern,
            lazy: async () => {
                const { SshCertificatesRoute } = await getLazyComponents();

                return {
                    element: (
                        <ModuleTitle title="SSH Certificates">
                            <SshCertificatesRoute />
                        </ModuleTitle>
                    ),
                };
            },
        },
    ],
} as const;
