import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useLicenseApi } from "~/system-settings/api/hooks";
import type {
    License_InstallOne_Req,
    License_InstallOne_Res,
    License_RemoveOne_Res,
} from "~/system-settings/api/services/license-services";
import { QK } from "~/system-settings/data/constants";

type InstallOneReq = License_InstallOne_Req["data"];
type InstallOneRes = License_InstallOne_Res;
type InstallOneOptions = Omit<UseMutationOptions<InstallOneRes, Error, InstallOneReq>, "mutationFn">;

function useInstallOne({ onSuccess, ...options }: InstallOneOptions = {}) {
    const { mutations } = useLicenseApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.installOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["system.license.find-one"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type RemoveOneRes = License_RemoveOne_Res;
type RemoveOneOptions = Omit<UseMutationOptions<RemoveOneRes>, "mutationFn">;

function useRemoveOne({ onSuccess, ...options }: RemoveOneOptions = {}) {
    const { mutations } = useLicenseApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.removeOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["system.license.find-one"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const LicenseCommands = Object.freeze({
    useInstallOne,
    useRemoveOne,
});
