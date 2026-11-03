import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useSshSettingsApi } from "~/organization/api/hooks";
import type {
    SshSettings_UpdateOne_Req,
    SshSettings_UpdateOne_Res,
} from "~/organization/api/services/ssh-settings-services";

import { QK as SHARED_QK } from "@application/shared/data/constants";

type UpdateOneReq = SshSettings_UpdateOne_Req["data"];
type UpdateOneRes = SshSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useSshSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            // The lifetime is read back from the organization itself.
            void queryClient.invalidateQueries({
                queryKey: [SHARED_QK["orgs.find-many"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const SshSettingsCommands = Object.freeze({
    useUpdateOne,
});
