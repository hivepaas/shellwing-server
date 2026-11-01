import { create } from "zustand";

import type { UpdateImPlatformStatusDialogOptions, UpdateImPlatformStatusDialogState } from "../types";

type State = UpdateImPlatformStatusDialogState & UpdateImPlatformStatusDialogOptions;

interface Actions {
    open: (id: string, options?: UpdateImPlatformStatusDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateImPlatformStatusDialogState = create<State & Actions>()(set => ({
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
