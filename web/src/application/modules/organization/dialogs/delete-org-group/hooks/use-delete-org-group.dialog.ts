import { type OrgGroup } from "~/organization/domain";

import { type DeleteOrgGroupDialogOptions } from "../types/delete-org-group.dialog.type";

import { useDeleteOrgGroupDialogState } from "./use-delete-org-group.dialog.state";

function createHook() {
    return function useDeleteOrgGroupDialog(props: DeleteOrgGroupDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useDeleteOrgGroupDialogState();

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

export const useDeleteOrgGroupDialog = createHook();
