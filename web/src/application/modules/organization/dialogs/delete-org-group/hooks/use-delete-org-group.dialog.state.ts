import { create } from "zustand";
import { type OrgGroup } from "~/organization/domain";

import {
    type DeleteOrgGroupDialogOptions,
    type DeleteOrgGroupDialogState,
} from "../types/delete-org-group.dialog.type";

type State = DeleteOrgGroupDialogState & DeleteOrgGroupDialogOptions;

interface Actions {
    open: (orgId: string, group: OrgGroup, options?: DeleteOrgGroupDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

const CLOSED: DeleteOrgGroupDialogState["state"] = {
    mode: "closed",
    orgId: null,
    group: null,
};

export const useDeleteOrgGroupDialogState = create<State & Actions>()(set => ({
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
