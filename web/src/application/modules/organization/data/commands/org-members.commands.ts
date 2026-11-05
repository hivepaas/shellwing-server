import { type QueryClient, type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useOrgMembersApi } from "~/organization/api/hooks";
import type {
    OrgMembers_UpdateRole_Req,
    OrgMembers_UpdateRole_Res,
    OrgMembers_UpdateStatus_Req,
    OrgMembers_UpdateStatus_Res,
} from "~/organization/api/services/org-members-services";
import { QK } from "~/organization/data/constants";
import { EOrgMemberStatus } from "~/organization/domain";

import { useProfileContext } from "@application/shared/context";
import { QK as SHARED_QK } from "@application/shared/data/constants";

/**
 * What a change of role or status shows in: the members, the accounts a removed member lost, the groups a removed
 * member was taken out of, and - when it is the signed-in user's own - their organizations, which decide the pages they
 * are offered.
 *
 * Who has just left is not asked for the members again: they may see them no more, and are taken off the page.
 */
function invalidateMembers(queryClient: QueryClient, { hasLeft = false }: { hasLeft?: boolean } = {}) {
    const refetchType = hasLeft ? "none" : "active";

    void queryClient.invalidateQueries({
        queryKey: [QK["organization.members.find-many"]],
        refetchType,
    });
    void queryClient.invalidateQueries({
        queryKey: [QK["organization.ssh-principals.find-one"]],
        refetchType,
    });
    void queryClient.invalidateQueries({
        queryKey: [QK["organization.groups.find-many"]],
        refetchType,
    });
    void queryClient.invalidateQueries({
        queryKey: [SHARED_QK["orgs.find-many"]],
    });
}

type UpdateRoleReq = OrgMembers_UpdateRole_Req["data"];
type UpdateRoleRes = OrgMembers_UpdateRole_Res;
type UpdateRoleOptions = Omit<UseMutationOptions<UpdateRoleRes, Error, UpdateRoleReq>, "mutationFn">;

function useUpdateRole({ onSuccess, ...options }: UpdateRoleOptions = {}) {
    const { mutations } = useOrgMembersApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateRole,
        onSuccess: (response, ...rest) => {
            invalidateMembers(queryClient);

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateStatusReq = OrgMembers_UpdateStatus_Req["data"];
type UpdateStatusRes = OrgMembers_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useOrgMembersApi();
    const queryClient = useQueryClient();
    const myUserId = useProfileContext(state => state.profile?.id);

    return useMutation({
        mutationFn: mutations.updateStatus,
        onSuccess: (response, request, ...rest) => {
            invalidateMembers(queryClient, {
                hasLeft: request.userId === myUserId && request.payload.status === EOrgMemberStatus.Removed,
            });

            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const OrgMembersCommands = Object.freeze({
    useUpdateRole,
    useUpdateStatus,
});
