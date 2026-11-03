import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useSshPrincipalsApi } from "~/organization/api/hooks";
import type {
    SshPrincipals_UpdateOne_Req,
    SshPrincipals_UpdateOne_Res,
} from "~/organization/api/services/ssh-principals-services";
import { QK } from "~/organization/data/constants";

import { QK as SHARED_QK } from "@application/shared/data/constants";

type UpdateOneReq = SshPrincipals_UpdateOne_Req["data"];
type UpdateOneRes = SshPrincipals_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useSshPrincipalsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["organization.members.find-many"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["organization.ssh-principals.find-one"]],
            });
            // The signed-in user's own accounts come with their organizations.
            void queryClient.invalidateQueries({
                queryKey: [SHARED_QK["orgs.find-many"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const SshPrincipalsCommands = Object.freeze({
    useUpdateOne,
});
