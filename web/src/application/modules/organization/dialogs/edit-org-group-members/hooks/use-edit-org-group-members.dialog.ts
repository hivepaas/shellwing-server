import { type OrgGroup } from "~/organization/domain";

import { type EditOrgGroupMembersDialogOptions } from "../types/edit-org-group-members.dialog.type";

import { useEditOrgGroupMembersDialogState } from "./use-edit-org-group-members.dialog.state";

function createHook() {
    return function useEditOrgGroupMembersDialog(props: EditOrgGroupMembersDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useEditOrgGroupMembersDialogState();

        return {
            state,
            actions: {
                open: (orgId: string, group: OrgGroup) => {
                    actions.open(orgId, group, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useEditOrgGroupMembersDialog = createHook();
