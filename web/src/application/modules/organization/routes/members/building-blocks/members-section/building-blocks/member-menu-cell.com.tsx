import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuTrigger,
} from "@components/ui/dropdown-menu";
import { LogOut, MoreVertical, Pause, Play, UserMinus, UserPlus } from "lucide-react";
import { toast } from "sonner";
import { OrgMembersCommands } from "~/organization/data/commands";
import { useRemoveOrgMemberDialog } from "~/organization/dialogs";
import { EOrgMemberStatus, type OrgMember } from "~/organization/domain";
import type { OrgMemberAbilities } from "~/organization/module-shared/utils";

import { EOrgRole } from "@application/shared/enums";

const DONE_MESSAGES: Partial<Record<string, string>> = {
    [EOrgMemberStatus.Suspended]: "suspended",
    [EOrgMemberStatus.Active]: "is active again",
};

/** Why some of what is listed is disabled, when it is not plain. */
function getNote(member: OrgMember, abilities: OrgMemberAbilities, actorIsOwner: boolean): string | null {
    if (abilities.isLastActiveOwner) {
        return "The only active owner. Make somebody else an owner first.";
    }

    if (!abilities.isSelf && member.role === EOrgRole.Owner && !actorIsOwner) {
        return "Only an owner can change an owner.";
    }

    return null;
}

function View({ orgId, member, abilities, actorIsOwner }: Props) {
    const [open, setOpen] = useState(false);
    const removeDialog = useRemoveOrgMemberDialog();
    const name = member.fullName || member.username;

    const { mutate: updateStatus, isPending } = OrgMembersCommands.useUpdateStatus({
        onSuccess: (_, request) => {
            const { status } = request.payload;

            if (status === EOrgMemberStatus.Active && member.status === EOrgMemberStatus.Removed) {
                toast.success(`${name} is back, as a member`);
            } else {
                toast.success(`${name} ${DONE_MESSAGES[status] ?? "updated"}`);
            }
        },
    });

    function setStatus(status: EOrgMemberStatus) {
        updateStatus({ orgId, userId: member.userId, payload: { status } });
    }

    function openRemove() {
        setOpen(false);
        removeDialog.actions.open(orgId, member, abilities.isSelf);
    }

    const note = getNote(member, abilities, actorIsOwner);

    return (
        // Not modal: a modal menu closing as the dialog it opens takes over can leave the page unclickable.
        <DropdownMenu
            open={open}
            onOpenChange={setOpen}
            modal={false}
        >
            <DropdownMenuTrigger asChild>
                <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8"
                    isLoading={isPending}
                >
                    {!isPending && <MoreVertical className="size-4" />}
                    <span className="sr-only">Actions for {member.username}</span>
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent
                align="end"
                className="max-w-[260px]"
            >
                {note !== null && (
                    <DropdownMenuLabel className="text-xs font-normal text-muted-foreground">{note}</DropdownMenuLabel>
                )}

                {abilities.isSelf && (
                    <DropdownMenuItem
                        variant="destructive"
                        disabled={!abilities.canLeave}
                        onSelect={openRemove}
                    >
                        <LogOut />
                        Leave organization
                    </DropdownMenuItem>
                )}

                {!abilities.isSelf && member.status === EOrgMemberStatus.Active && (
                    <DropdownMenuItem
                        disabled={!abilities.canSuspend}
                        onSelect={() => {
                            setStatus(EOrgMemberStatus.Suspended);
                        }}
                    >
                        <Pause />
                        Suspend
                    </DropdownMenuItem>
                )}

                {!abilities.isSelf && member.status === EOrgMemberStatus.Suspended && (
                    <DropdownMenuItem
                        disabled={!abilities.canReactivate}
                        onSelect={() => {
                            setStatus(EOrgMemberStatus.Active);
                        }}
                    >
                        <Play />
                        Reactivate
                    </DropdownMenuItem>
                )}

                {!abilities.isSelf && member.status !== EOrgMemberStatus.Removed && (
                    <DropdownMenuItem
                        variant="destructive"
                        disabled={!abilities.canRemove}
                        onSelect={openRemove}
                    >
                        <UserMinus />
                        Remove from organization
                    </DropdownMenuItem>
                )}

                {!abilities.isSelf && member.status === EOrgMemberStatus.Removed && (
                    <DropdownMenuItem
                        disabled={!abilities.canAddBack}
                        onSelect={() => {
                            setStatus(EOrgMemberStatus.Active);
                        }}
                    >
                        <UserPlus />
                        Add back
                    </DropdownMenuItem>
                )}
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

interface Props {
    orgId: string;
    member: OrgMember;
    abilities: OrgMemberAbilities;
    actorIsOwner: boolean;
}

export const MemberMenuCell = memo(View);
