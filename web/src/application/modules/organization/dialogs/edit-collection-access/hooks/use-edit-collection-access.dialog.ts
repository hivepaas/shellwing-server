import { type Collection } from "~/organization/domain";

import { type EditCollectionAccessDialogOptions } from "../types/edit-collection-access.dialog.type";

import { useEditCollectionAccessDialogState } from "./use-edit-collection-access.dialog.state";

function createHook() {
    return function useEditCollectionAccessDialog(props: EditCollectionAccessDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useEditCollectionAccessDialogState();

        return {
            state,
            actions: {
                open: (orgId: string, collection: Collection) => {
                    actions.open(orgId, collection, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useEditCollectionAccessDialog = createHook();
