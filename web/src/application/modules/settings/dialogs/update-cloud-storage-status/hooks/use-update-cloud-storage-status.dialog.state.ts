import { create } from "zustand";

import type { UpdateCloudStorageStatusDialogOptions, UpdateCloudStorageStatusDialogState } from "../types";

type State = UpdateCloudStorageStatusDialogState & UpdateCloudStorageStatusDialogOptions;

interface Actions {
    open: (id: string, options?: UpdateCloudStorageStatusDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateCloudStorageStatusDialogState = create<State & Actions>()(set => ({
    state: { mode: "closed" },
    props: {},
    open: (id, options = {}) => {
        set({ state: { mode: "open", id }, ...options });
    },
    close: () => {
        set({ state: { mode: "closed" } });
    },
    clear: () => {
        set({ props: {} });
    },
    destroy: () => {
        set(state => {
            if (state.state.mode === "closed") return state;
            return { state: { mode: "closed" }, props: {} };
        });
    },
}));
