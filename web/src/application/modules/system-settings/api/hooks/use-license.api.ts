import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SystemSettingsApiContext } from "~/system-settings/api/api-context";
import type { License_InstallOne_Req } from "~/system-settings/api/services/license-services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useLicenseApi() {
        const { api } = use(SystemSettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findOne: async (signal?: AbortSignal) => {
                    const result = await api.system.license.findOne(signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get the license",
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
                installOne: async (data: License_InstallOne_Req["data"]) => {
                    const result = await api.system.license.installOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to install the license key",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                removeOne: async () => {
                    const result = await api.system.license.removeOne();

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to remove the license key",
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

export const useLicenseApi = createHook();
