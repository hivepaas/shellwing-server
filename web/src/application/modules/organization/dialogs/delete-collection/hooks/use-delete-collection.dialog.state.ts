import { create } from "zustand";
import { type Collection } from "~/organization/domain";

import {
    type DeleteCollectionDialogOptions,
    type DeleteCollectionDialogState,
} from "../types/delete-collection.dialog.type";

type State = DeleteCollectionDialogState & DeleteCollectionDialogOptions;

interface Actions {
    open: (orgId: string, collection: Collection, options?: DeleteCollectionDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

const CLOSED: DeleteCollectionDialogState["state"] = {
    mode: "closed",
    orgId: null,
    collection: null,
};

export const useDeleteCollectionDialogState = create<State & Actions>()(set => ({
    state: CLOSED,
    props: {},

    open: (orgId, collection, options = {}) => {
        set({
            state: {
                mode: "open",
                orgId,
                collection,
            },
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
