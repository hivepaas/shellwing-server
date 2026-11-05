import { type OrgGroup } from "~/organization/domain";

import { type EditOrgGroupDialogOptions } from "../types/edit-org-group.dialog.type";

import { useEditOrgGroupDialogState } from "./use-edit-org-group.dialog.state";

function createHook() {
    return function useEditOrgGroupDialog(props: EditOrgGroupDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useEditOrgGroupDialogState();

        return {
            state,
            actions: {
                create: (orgId: string) => {
                    actions.open(orgId, null, { props });
                },
                edit: (orgId: string, group: OrgGroup) => {
                    actions.open(orgId, group, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useEditOrgGroupDialog = createHook();
