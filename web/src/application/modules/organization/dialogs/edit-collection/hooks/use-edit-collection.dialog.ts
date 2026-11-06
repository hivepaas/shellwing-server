import { type Collection } from "~/organization/domain";

import { type EditCollectionDialogOptions } from "../types/edit-collection.dialog.type";

import { useEditCollectionDialogState } from "./use-edit-collection.dialog.state";

function createHook() {
    return function useEditCollectionDialog(props: EditCollectionDialogOptions["props"] = {}) {
        const { state, props: _, ...actions } = useEditCollectionDialogState();

        return {
            state,
            actions: {
                create: (orgId: string) => {
                    actions.open(orgId, null, { props });
                },
                edit: (orgId: string, collection: Collection) => {
                    actions.open(orgId, collection, { props });
                },
                close: () => {
                    actions.close();
                },
            },
        };
    };
}

export const useEditCollectionDialog = createHook();
