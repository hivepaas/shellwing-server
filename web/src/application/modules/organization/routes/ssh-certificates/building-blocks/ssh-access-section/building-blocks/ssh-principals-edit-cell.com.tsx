import { memo } from "react";

import { Button } from "@components/ui/button";
import { Pencil } from "lucide-react";
import { useEditSshPrincipalsDialog } from "~/organization/dialogs";
import type { OrgMember } from "~/organization/domain";

function View({ orgId, member }: Props) {
    const editDialog = useEditSshPrincipalsDialog();

    return (
        <Button
            type="button"
            variant="ghost"
            size="icon"
            className="h-8 w-8"
            aria-label={`Edit the accounts of ${member.username}`}
            onClick={() => {
                editDialog.actions.open(orgId, member);
            }}
        >
            <Pencil className="size-4" />
        </Button>
    );
}

interface Props {
    orgId: string;
    member: OrgMember;
}

export const SshPrincipalsEditCell = memo(View);
