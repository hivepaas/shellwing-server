import { create } from "zustand";
import { type OrgGroup } from "~/organization/domain";

import { type EditOrgGroupDialogOptions, type EditOrgGroupDialogState } from "../types/edit-org-group.dialog.type";

type State = EditOrgGroupDialogState & EditOrgGroupDialogOptions;

interface Actions {
    /** Make a group, or - given one - edit it. */
    open: (orgId: string, group: OrgGroup | null, options?: EditOrgGroupDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

const CLOSED: EditOrgGroupDialogState["state"] = {
    mode: "closed",
    orgId: null,
    group: null,
};

export const useEditOrgGroupDialogState = create<State & Actions>()(set => ({
    state: CLOSED,
    props: {},

    open: (orgId, group, options = {}) => {
        set({
            state: group === null ? { mode: "create", orgId, group } : { mode: "edit", orgId, group },
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
