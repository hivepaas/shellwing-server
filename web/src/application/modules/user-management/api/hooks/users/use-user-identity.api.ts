import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { UsersApiContext } from "~/user-management/api/api-context";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useUserIdentityApi() {
        const { api } = use(UsersApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                /**
                 * Find the signed-in person's identity
                 */
                findMine: async (signal?: AbortSignal) => {
                    const result = await api.users.identity.findMine(signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get your encryption key",
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
        };
    };
}

export const useUserIdentityApi = createHook();
