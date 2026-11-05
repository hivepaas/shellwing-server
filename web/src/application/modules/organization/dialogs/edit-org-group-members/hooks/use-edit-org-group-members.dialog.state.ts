import { create } from "zustand";
import { type OrgGroup } from "~/organization/domain";

import {
    type EditOrgGroupMembersDialogOptions,
    type EditOrgGroupMembersDialogState,
} from "../types/edit-org-group-members.dialog.type";

type State = EditOrgGroupMembersDialogState & EditOrgGroupMembersDialogOptions;

interface Actions {
    open: (orgId: string, group: OrgGroup, options?: EditOrgGroupMembersDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

const CLOSED: EditOrgGroupMembersDialogState["state"] = {
    mode: "closed",
    orgId: null,
    group: null,
};

export const useEditOrgGroupMembersDialogState = create<State & Actions>()(set => ({
    state: CLOSED,
    props: {},

    open: (orgId, group, options = {}) => {
        set({
            state: {
                mode: "open",
                orgId,
                group,
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
