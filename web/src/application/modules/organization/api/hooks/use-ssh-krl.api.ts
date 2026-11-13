import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { OrganizationApiContext } from "~/organization/api/api-context";
import type {
    SshKrlToken_CreateOne_Req,
    SshKrlToken_DeleteOne_Req,
} from "~/organization/api/services/ssh-krl-services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useSshKrlApi() {
        const { api } = use(OrganizationApiContext);
        const { notifyError } = useApiErrorNotifications();

        const mutations = useMemo(
            () => ({
                createOne: async (data: SshKrlToken_CreateOne_Req["data"]) => {
                    const result = await api.organization.ssh.krl.createOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to make a revocation list token",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: SshKrlToken_DeleteOne_Req["data"]) => {
                    const result = await api.organization.ssh.krl.deleteOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to take the revocation list token away",
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
            mutations,
        };
    };
}

export const useSshKrlApi = createHook();
