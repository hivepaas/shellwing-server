import { type OrgMember } from "~/organization/domain";

export interface EditSshPrincipalsDialogState {
    state:
        | {
              mode: "open";
              orgId: string;
              member: OrgMember;
          }
        | {
              mode: "closed";
              orgId: null;
              member: null;
          };
}

export interface EditSshPrincipalsDialogOptions {
    props?: {
        onSuccess?: (principals: string[]) => void;
        onClose?: () => void;
    };
}
