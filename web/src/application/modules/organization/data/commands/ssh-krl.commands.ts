import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useSshKrlApi } from "~/organization/api/hooks";
import type {
    SshKrlToken_CreateOne_Req,
    SshKrlToken_CreateOne_Res,
    SshKrlToken_DeleteOne_Req,
    SshKrlToken_DeleteOne_Res,
} from "~/organization/api/services/ssh-krl-services";

import { QK as SHARED_QK } from "@application/shared/data/constants";

type CreateOneReq = SshKrlToken_CreateOne_Req["data"];
type CreateOneRes = SshKrlToken_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useSshKrlApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            // Whether there is a token, and since when, is read back from the organization itself.
            void queryClient.invalidateQueries({
                queryKey: [SHARED_QK["orgs.find-many"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = SshKrlToken_DeleteOne_Req["data"];
type DeleteOneRes = SshKrlToken_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useSshKrlApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [SHARED_QK["orgs.find-many"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const SshKrlCommands = Object.freeze({
    useCreateOne,
    useDeleteOne,
});
