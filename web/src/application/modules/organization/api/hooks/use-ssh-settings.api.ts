import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { OrganizationApiContext } from "~/organization/api/api-context";
import type { SshSettings_UpdateOne_Req } from "~/organization/api/services/ssh-settings-services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useSshSettingsApi() {
        const { api } = use(OrganizationApiContext);
        const { notifyError } = useApiErrorNotifications();

        const mutations = useMemo(
            () => ({
                updateOne: async (data: SshSettings_UpdateOne_Req["data"]) => {
                    const result = await api.organization.ssh.settings.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to save the certificate lifetime",
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

export const useSshSettingsApi = createHook();
