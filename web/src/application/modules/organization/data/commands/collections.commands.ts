import { type QueryClient, type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useCollectionsApi } from "~/organization/api/hooks";
import type {
    Collections_CreateOne_Req,
    Collections_CreateOne_Res,
    Collections_DeleteOne_Req,
    Collections_DeleteOne_Res,
    Collections_FindOne_Res,
    Collections_SetGrants_Req,
    Collections_SetGrants_Res,
    Collections_UpdateOne_Req,
    Collections_UpdateOne_Res,
} from "~/organization/api/services/collections-services";
import { QK } from "~/organization/data/constants";

/**
 * What a change to a collection shows in: the collections, and the collection itself with who has access.
 */
function invalidateCollections(queryClient: QueryClient) {
    void queryClient.invalidateQueries({
        queryKey: [QK["organization.collections.find-many"]],
    });
    void queryClient.invalidateQueries({
        queryKey: [QK["organization.collections.find-one"]],
    });
}

type CreateOneReq = Collections_CreateOne_Req["data"];
type CreateOneRes = Collections_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useCollectionsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            invalidateCollections(queryClient);

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = Collections_UpdateOne_Req["data"];
type UpdateOneRes = Collections_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useCollectionsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            invalidateCollections(queryClient);

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = Collections_DeleteOne_Req["data"];
type DeleteOneRes = Collections_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useCollectionsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["organization.collections.find-many"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type SetGrantsReq = Collections_SetGrants_Req["data"];
type SetGrantsRes = Collections_SetGrants_Res;
type SetGrantsOptions = Omit<UseMutationOptions<SetGrantsRes, Error, SetGrantsReq>, "mutationFn">;

function useSetGrants({ onSuccess, ...options }: SetGrantsOptions = {}) {
    const { mutations } = useCollectionsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.setGrants,
        onSuccess: (response, request, ...rest) => {
            // The grants are known at once; who has access, and so who waits for the key, comes with the next read.
            queryClient.setQueryData<Collections_FindOne_Res>(
                [QK["organization.collections.find-one"], { orgId: request.orgId, collectionId: request.collectionId }],
                previous => (previous ? { ...previous, data: { ...previous.data, grants: response.data } } : previous),
            );
            invalidateCollections(queryClient);

            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const CollectionsCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useDeleteOne,
    useSetGrants,
});
