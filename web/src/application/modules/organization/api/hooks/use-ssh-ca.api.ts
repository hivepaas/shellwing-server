import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { OrganizationApiContext } from "~/organization/api/api-context";
import type {
    SshCas_CreateOne_Req,
    SshCas_DeleteOne_Req,
    SshCas_FindMany_Req,
    SshCas_SetStatus_Req,
} from "~/organization/api/services/ssh-ca-services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useSshCaApi() {
        const { api } = use(OrganizationApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findMany: async (data: SshCas_FindMany_Req["data"], signal?: AbortSignal) => {
                    const result = await api.organization.ssh.cas.findMany({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get SSH certificate authorities",
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
                createOne: async (data: SshCas_CreateOne_Req["data"]) => {
                    const result = await api.organization.ssh.cas.createOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create SSH certificate authority",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                setStatus: async (data: SshCas_SetStatus_Req["data"]) => {
                    const result = await api.organization.ssh.cas.setStatus({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to change the SSH certificate authority",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: SshCas_DeleteOne_Req["data"]) => {
                    const result = await api.organization.ssh.cas.deleteOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete SSH certificate authority",
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

export const useSshCaApi = createHook();
