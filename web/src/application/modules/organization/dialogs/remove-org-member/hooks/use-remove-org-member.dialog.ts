import { type OrgMember } from "~/organization/domain";

import { type RemoveOrgMemberDialogOptions } from "../types/remove-org-member.dialog.type";

import { useRemoveOrgMemberDialogState } from "./use-remove-org-member.dialog.state";

function createHook() {
    return function useRemoveOrgMemberDialog(props: RemoveOrgMemberDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useRemoveOrgMemberDialogState();

        return {
            state,
            actions: {
                /** Remove a member, or - the signed-in user themselves - leave. */
                open: (orgId: string, member: OrgMember, isSelf: boolean) => {
                    actions.open(orgId, member, isSelf, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useRemoveOrgMemberDialog = createHook();
