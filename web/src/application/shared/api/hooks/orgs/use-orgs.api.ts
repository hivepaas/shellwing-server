import { use, useMemo } from "react";

import { match } from "oxide.ts";

import { ApplicationApiContext } from "@application/shared/api/api-context";

function createHook() {
    return function useOrgsApi() {
        const { api } = use(ApplicationApiContext);

        const queries = useMemo(
            () => ({
                /**
                 * The organizations the signed-in user is a member of.
                 *
                 * Not notified: the sidebar asks for them on every page, to know whether to offer the
                 * organization's admin pages, and a failure there is no reason to interrupt. The pages
                 * that need them show the failure themselves.
                 */
                findMany: async (signal?: AbortSignal) => {
                    const result = await api.orgs.findMany(signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            throw error;
                        },
                    });
                },
            }),
            [api],
        );

        return {
            queries,
        };
    };
}

export const useOrgsApi = createHook();
