import { create } from "zustand";
import { type OrgMember } from "~/organization/domain";

import {
    type RemoveOrgMemberDialogOptions,
    type RemoveOrgMemberDialogState,
} from "../types/remove-org-member.dialog.type";

type State = RemoveOrgMemberDialogState & RemoveOrgMemberDialogOptions;

interface Actions {
    open: (orgId: string, member: OrgMember, isSelf: boolean, options?: RemoveOrgMemberDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

const CLOSED: RemoveOrgMemberDialogState["state"] = {
    mode: "closed",
    orgId: null,
    member: null,
    isSelf: false,
};

export const useRemoveOrgMemberDialogState = create<State & Actions>()(set => ({
    state: CLOSED,
    props: {},

    open: (orgId, member, isSelf, options = {}) => {
        set({
            state: {
                mode: "open",
                orgId,
                member,
                isSelf,
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
