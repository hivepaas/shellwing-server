import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useSshCaApi } from "~/organization/api/hooks";
import type {
    SshCas_CreateOne_Req,
    SshCas_CreateOne_Res,
    SshCas_DeleteOne_Req,
    SshCas_DeleteOne_Res,
    SshCas_SetStatus_Req,
    SshCas_SetStatus_Res,
} from "~/organization/api/services/ssh-ca-services";
import { QK } from "~/organization/data/constants";

type CreateOneReq = SshCas_CreateOne_Req["data"];
type CreateOneRes = SshCas_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useSshCaApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            // A new one is published beside the one that signs, or replaces it: either way the list changed.
            void queryClient.invalidateQueries({
                queryKey: [QK["organization.ssh-cas.find-many"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = SshCas_DeleteOne_Req["data"];
type DeleteOneRes = SshCas_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useSshCaApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["organization.ssh-cas.find-many"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type SetStatusReq = SshCas_SetStatus_Req["data"];
type SetStatusRes = SshCas_SetStatus_Res;
type SetStatusOptions = Omit<UseMutationOptions<SetStatusRes, Error, SetStatusReq>, "mutationFn">;

function useSetStatus({ onSuccess, ...options }: SetStatusOptions = {}) {
    const { mutations } = useSshCaApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.setStatus,
        onSuccess: (response, ...rest) => {
            // Promoting one demotes the other, so both rows moved.
            void queryClient.invalidateQueries({
                queryKey: [QK["organization.ssh-cas.find-many"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const SshCaCommands = Object.freeze({
    useCreateOne,
    useSetStatus,
    useDeleteOne,
});
