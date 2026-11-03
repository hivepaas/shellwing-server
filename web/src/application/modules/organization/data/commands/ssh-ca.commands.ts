import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useSshCaApi } from "~/organization/api/hooks";
import type {
    SshCas_CreateOne_Req,
    SshCas_CreateOne_Res,
    SshCas_DeleteOne_Req,
    SshCas_DeleteOne_Res,
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
            // The one active before is retired by it.
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

export const SshCaCommands = Object.freeze({
    useCreateOne,
    useDeleteOne,
});
