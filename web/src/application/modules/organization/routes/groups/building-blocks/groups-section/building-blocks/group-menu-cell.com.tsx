import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, Pencil, Trash2, UsersRound } from "lucide-react";
import { useDeleteOrgGroupDialog, useEditOrgGroupDialog, useEditOrgGroupMembersDialog } from "~/organization/dialogs";
import type { OrgGroup } from "~/organization/domain";

function View({ orgId, group }: Props) {
    const [open, setOpen] = useState(false);
    const editDialog = useEditOrgGroupDialog();
    const membersDialog = useEditOrgGroupMembersDialog();
    const deleteDialog = useDeleteOrgGroupDialog();

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
                >
                    <MoreVertical className="size-4" />
                    <span className="sr-only">Actions for {group.name}</span>
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
                <DropdownMenuItem
                    onSelect={() => {
                        setOpen(false);
                        editDialog.actions.edit(orgId, group);
                    }}
                >
                    <Pencil />
                    Edit
                </DropdownMenuItem>
                <DropdownMenuItem
                    onSelect={() => {
                        setOpen(false);
                        membersDialog.actions.open(orgId, group);
                    }}
                >
                    <UsersRound />
                    Members
                </DropdownMenuItem>
                <DropdownMenuItem
                    variant="destructive"
                    onSelect={() => {
                        setOpen(false);
                        deleteDialog.actions.open(orgId, group);
                    }}
                >
                    <Trash2 />
                    Delete
                </DropdownMenuItem>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

interface Props {
    orgId: string;
    group: OrgGroup;
}

export const GroupMenuCell = memo(View);
