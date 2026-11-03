import { create } from "zustand";
import { type OrgMember } from "~/organization/domain";

import {
    type EditSshPrincipalsDialogOptions,
    type EditSshPrincipalsDialogState,
} from "../types/edit-ssh-principals.dialog.type";

type State = EditSshPrincipalsDialogState & EditSshPrincipalsDialogOptions;

interface Actions {
    open: (orgId: string, member: OrgMember, options?: EditSshPrincipalsDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useEditSshPrincipalsDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
        orgId: null,
        member: null,
    },
    props: {},

    open: (orgId, member, options = {}) => {
        set({
            state: {
                mode: "open",
                orgId,
                member,
            },
            ...options,
        });
    },

    close: () => {
        set({
            state: {
                mode: "closed",
                orgId: null,
                member: null,
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
                    member: null,
                },
                props: {},
            };
        });
    },
}));
