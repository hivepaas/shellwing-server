import { create } from "zustand";
import { type Collection } from "~/organization/domain";

import { type EditCollectionDialogOptions, type EditCollectionDialogState } from "../types/edit-collection.dialog.type";

type State = EditCollectionDialogState & EditCollectionDialogOptions;

interface Actions {
    /** Make a collection, or - given one - edit it. */
    open: (orgId: string, collection: Collection | null, options?: EditCollectionDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

const CLOSED: EditCollectionDialogState["state"] = {
    mode: "closed",
    orgId: null,
    collection: null,
};

export const useEditCollectionDialogState = create<State & Actions>()(set => ({
    state: CLOSED,
    props: {},

    open: (orgId, collection, options = {}) => {
        set({
            state: collection === null ? { mode: "create", orgId, collection } : { mode: "edit", orgId, collection },
            ...options,
        });
    },

    close: () => {
        set({
            state: CLOSED,
        });
    },

    clear: () => {
        set({
            props: {},
        });
    },

    destroy: () => {
        set(state => {
            if (state.state.mode === "closed") {
                return state;
            }

            return {
                state: CLOSED,
                props: {},
            };
        });
    },
}));
