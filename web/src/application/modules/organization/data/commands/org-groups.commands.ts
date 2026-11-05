import { type QueryClient, type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useOrgGroupsApi } from "~/organization/api/hooks";
import type {
    OrgGroups_CreateOne_Req,
    OrgGroups_CreateOne_Res,
    OrgGroups_DeleteOne_Req,
    OrgGroups_DeleteOne_Res,
    OrgGroups_SetMembers_Req,
    OrgGroups_SetMembers_Res,
    OrgGroups_UpdateOne_Req,
    OrgGroups_UpdateOne_Res,
} from "~/organization/api/services/org-groups-services";
import { QK } from "~/organization/data/constants";

import { QK as SHARED_QK } from "@application/shared/data/constants";

/**
 * What a change to a group shows in: the groups, the members with their groups (the members' and the SSH certificates'
 * pages), and the signed-in user's organizations, whose accounts include their groups'.
 */
function invalidateGroups(queryClient: QueryClient) {
    void queryClient.invalidateQueries({
        queryKey: [QK["organization.groups.find-many"]],
    });
    void queryClient.invalidateQueries({
        queryKey: [QK["organization.members.find-many"]],
    });
    void queryClient.invalidateQueries({
        queryKey: [SHARED_QK["orgs.find-many"]],
    });
}

type CreateOneReq = OrgGroups_CreateOne_Req["data"];
type CreateOneRes = OrgGroups_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useOrgGroupsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            // A new group has no members: nobody's accounts change yet.
            void queryClient.invalidateQueries({
                queryKey: [QK["organization.groups.find-many"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = OrgGroups_UpdateOne_Req["data"];
type UpdateOneRes = OrgGroups_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useOrgGroupsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            invalidateGroups(queryClient);

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = OrgGroups_DeleteOne_Req["data"];
type DeleteOneRes = OrgGroups_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useOrgGroupsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            invalidateGroups(queryClient);

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type SetMembersReq = OrgGroups_SetMembers_Req["data"];
type SetMembersRes = OrgGroups_SetMembers_Res;
type SetMembersOptions = Omit<UseMutationOptions<SetMembersRes, Error, SetMembersReq>, "mutationFn">;

function useSetMembers({ onSuccess, ...options }: SetMembersOptions = {}) {
    const { mutations } = useOrgGroupsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.setMembers,
        onSuccess: (response, ...rest) => {
            invalidateGroups(queryClient);

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const OrgGroupsCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useDeleteOne,
    useSetMembers,
});
