import { create } from "zustand";

import type { UpdateEmailAccountStatusDialogOptions, UpdateEmailAccountStatusDialogState } from "../types";

type State = UpdateEmailAccountStatusDialogState & UpdateEmailAccountStatusDialogOptions;

interface Actions {
    open: (id: string, options?: UpdateEmailAccountStatusDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateEmailAccountStatusDialogState = create<State & Actions>()(set => ({
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
