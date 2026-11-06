import { type Collection } from "~/organization/domain";

import { type DeleteCollectionDialogOptions } from "../types/delete-collection.dialog.type";

import { useDeleteCollectionDialogState } from "./use-delete-collection.dialog.state";

function createHook() {
    return function useDeleteCollectionDialog(props: DeleteCollectionDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useDeleteCollectionDialogState();

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

export const useDeleteCollectionDialog = createHook();
