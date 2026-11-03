import { type ReplaceSshCaDialogOptions } from "../types/replace-ssh-ca.dialog.type";

import { useReplaceSshCaDialogState } from "./use-replace-ssh-ca.dialog.state";

function createHook() {
    return function useReplaceSshCaDialog(props: ReplaceSshCaDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useReplaceSshCaDialogState();

        return {
            state,
            actions: {
                open: (orgId: string) => {
                    actions.open(orgId, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useReplaceSshCaDialog = createHook();
