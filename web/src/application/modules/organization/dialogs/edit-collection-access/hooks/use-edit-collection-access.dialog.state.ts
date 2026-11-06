import { create } from "zustand";
import { type Collection } from "~/organization/domain";

import {
    type EditCollectionAccessDialogOptions,
    type EditCollectionAccessDialogState,
} from "../types/edit-collection-access.dialog.type";

type State = EditCollectionAccessDialogState & EditCollectionAccessDialogOptions;

interface Actions {
    open: (orgId: string, collection: Collection, options?: EditCollectionAccessDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

const CLOSED: EditCollectionAccessDialogState["state"] = {
    mode: "closed",
    orgId: null,
    collection: null,
};

export const useEditCollectionAccessDialogState = create<State & Actions>()(set => ({
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
