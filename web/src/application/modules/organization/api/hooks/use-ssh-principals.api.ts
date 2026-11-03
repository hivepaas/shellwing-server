import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { OrganizationApiContext } from "~/organization/api/api-context";
import type {
    SshPrincipals_FindOne_Req,
    SshPrincipals_UpdateOne_Req,
} from "~/organization/api/services/ssh-principals-services";

import { isSshPrincipalInvalidException, useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useSshPrincipalsApi() {
        const { api } = use(OrganizationApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findOne: async (data: SshPrincipals_FindOne_Req["data"], signal?: AbortSignal) => {
                    const result = await api.organization.ssh.principals.findOne({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get the accounts of the member",
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
                updateOne: async (data: SshPrincipals_UpdateOne_Req["data"]) => {
                    const result = await api.organization.ssh.principals.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            // A name refused is shown beside the names, in the form they were typed in.
                            if (!isSshPrincipalInvalidException(error)) {
                                notifyError({
                                    message: "Failed to save the accounts of the member",
                                    error,
                                });
                            }

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

export const useSshPrincipalsApi = createHook();
