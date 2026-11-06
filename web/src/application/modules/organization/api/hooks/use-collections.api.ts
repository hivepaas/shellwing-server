import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { OrganizationApiContext } from "~/organization/api/api-context";
import type {
    Collections_CreateOne_Req,
    Collections_DeleteOne_Req,
    Collections_FindMany_Req,
    Collections_FindOne_Req,
    Collections_SetGrants_Req,
    Collections_UpdateOne_Req,
} from "~/organization/api/services/collections-services";

import { isAlreadyExistException, useApiErrorNotifications } from "@infrastructure/api";

/** A name taken is shown beside it in the form it was typed in. */
function isShownInForm(error: Error): boolean {
    return isAlreadyExistException(error);
}

function createHook() {
    return function useCollectionsApi() {
        const { api } = use(OrganizationApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findMany: async (data: Collections_FindMany_Req["data"], signal?: AbortSignal) => {
                    const result = await api.organization.collections.findMany({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get the collections of the organization",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findOne: async (data: Collections_FindOne_Req["data"], signal?: AbortSignal) => {
                    const result = await api.organization.collections.findOne({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get the collection",
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
                createOne: async (data: Collections_CreateOne_Req["data"]) => {
                    const result = await api.organization.collections.createOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            if (!isShownInForm(error)) {
                                notifyError({
                                    message: "Failed to create the collection",
                                    error,
                                });
                            }

                            throw error;
                        },
                    });
                },
                updateOne: async (data: Collections_UpdateOne_Req["data"]) => {
                    const result = await api.organization.collections.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            if (!isShownInForm(error)) {
                                notifyError({
                                    message: "Failed to save the collection",
                                    error,
                                });
                            }

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: Collections_DeleteOne_Req["data"]) => {
                    const result = await api.organization.collections.deleteOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete the collection",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                setGrants: async (data: Collections_SetGrants_Req["data"]) => {
                    const result = await api.organization.collections.setGrants({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to save the access to the collection",
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

export const useCollectionsApi = createHook();
