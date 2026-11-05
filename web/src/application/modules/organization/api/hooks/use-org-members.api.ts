import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { OrganizationApiContext } from "~/organization/api/api-context";
import type {
    OrgMembers_FindMany_Req,
    OrgMembers_UpdateRole_Req,
    OrgMembers_UpdateStatus_Req,
} from "~/organization/api/services/org-members-services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useOrgMembersApi() {
        const { api } = use(OrganizationApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findMany: async (data: OrgMembers_FindMany_Req["data"], signal?: AbortSignal) => {
                    const result = await api.organization.members.findMany({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get the members of the organization",
                                error,
                            });

                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                updateRole: async (data: OrgMembers_UpdateRole_Req["data"]) => {
                    const result = await api.organization.members.updateRole({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to change the role",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateStatus: async (data: OrgMembers_UpdateStatus_Req["data"]) => {
                    const result = await api.organization.members.updateStatus({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to change the status",
                                error,
                            });

                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        return {
            queries,
            mutations,
        };
    };
}

export const useOrgMembersApi = createHook();
