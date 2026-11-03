import { type OrgMember } from "~/organization/domain";

import { type EditSshPrincipalsDialogOptions } from "../types/edit-ssh-principals.dialog.type";

import { useEditSshPrincipalsDialogState } from "./use-edit-ssh-principals.dialog.state";

function createHook() {
    return function useEditSshPrincipalsDialog(props: EditSshPrincipalsDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useEditSshPrincipalsDialogState();

        return {
            state,
            actions: {
                open: (orgId: string, member: OrgMember) => {
                    actions.open(orgId, member, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useEditSshPrincipalsDialog = createHook();
