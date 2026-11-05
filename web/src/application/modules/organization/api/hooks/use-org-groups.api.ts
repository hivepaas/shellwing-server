import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { OrganizationApiContext } from "~/organization/api/api-context";
import type {
    OrgGroups_CreateOne_Req,
    OrgGroups_DeleteOne_Req,
    OrgGroups_FindMany_Req,
    OrgGroups_SetMembers_Req,
    OrgGroups_UpdateOne_Req,
} from "~/organization/api/services/org-groups-services";

import { isAlreadyExistException, isSshPrincipalInvalidException, useApiErrorNotifications } from "@infrastructure/api";

/** A name taken, or an account refused, is shown beside it in the form it was typed in. */
function isShownInForm(error: Error): boolean {
    return isAlreadyExistException(error) || isSshPrincipalInvalidException(error);
}

function createHook() {
    return function useOrgGroupsApi() {
        const { api } = use(OrganizationApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findMany: async (data: OrgGroups_FindMany_Req["data"], signal?: AbortSignal) => {
                    const result = await api.organization.groups.findMany({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get the groups of the organization",
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
                createOne: async (data: OrgGroups_CreateOne_Req["data"]) => {
                    const result = await api.organization.groups.createOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            if (!isShownInForm(error)) {
                                notifyError({
                                    message: "Failed to create the group",
                                    error,
                                });
                            }

                            throw error;
                        },
                    });
                },
                updateOne: async (data: OrgGroups_UpdateOne_Req["data"]) => {
                    const result = await api.organization.groups.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            if (!isShownInForm(error)) {
                                notifyError({
                                    message: "Failed to save the group",
                                    error,
                                });
                            }

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: OrgGroups_DeleteOne_Req["data"]) => {
                    const result = await api.organization.groups.deleteOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete the group",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                setMembers: async (data: OrgGroups_SetMembers_Req["data"]) => {
                    const result = await api.organization.groups.setMembers({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to save the members of the group",
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

export const useOrgGroupsApi = createHook();
