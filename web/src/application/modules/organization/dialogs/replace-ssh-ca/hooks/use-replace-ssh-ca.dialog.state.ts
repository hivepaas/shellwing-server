import { create } from "zustand";

import { type ReplaceSshCaDialogOptions, type ReplaceSshCaDialogState } from "../types/replace-ssh-ca.dialog.type";

type State = ReplaceSshCaDialogState & ReplaceSshCaDialogOptions;

interface Actions {
    open: (orgId: string, options?: ReplaceSshCaDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useReplaceSshCaDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
        orgId: null,
    },
    props: {},

    open: (orgId, options = {}) => {
        set({
            state: {
                mode: "open",
                orgId,
            },
            ...options,
        });
    },

    close: () => {
        set({
            state: {
                mode: "closed",
                orgId: null,
            },
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
                state: {
                    mode: "closed",
                    orgId: null,
                },
                props: {},
            };
        });
    },
}));
