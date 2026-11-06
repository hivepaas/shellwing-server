import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { KeyRound, MoreVertical, Pencil, Trash2 } from "lucide-react";
import {
    useDeleteCollectionDialog,
    useEditCollectionAccessDialog,
    useEditCollectionDialog,
} from "~/organization/dialogs";
import type { Collection } from "~/organization/domain";

function View({ orgId, collection }: Props) {
    const [open, setOpen] = useState(false);
    const editDialog = useEditCollectionDialog();
    const accessDialog = useEditCollectionAccessDialog();
    const deleteDialog = useDeleteCollectionDialog();

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
                    <span className="sr-only">Actions for {collection.name}</span>
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
                <DropdownMenuItem
                    onSelect={() => {
                        setOpen(false);
                        editDialog.actions.edit(orgId, collection);
                    }}
                >
                    <Pencil />
                    Edit
                </DropdownMenuItem>
                <DropdownMenuItem
                    onSelect={() => {
                        setOpen(false);
                        accessDialog.actions.open(orgId, collection);
                    }}
                >
                    <KeyRound />
                    Access
                </DropdownMenuItem>
                <DropdownMenuItem
                    variant="destructive"
                    onSelect={() => {
                        setOpen(false);
                        deleteDialog.actions.open(orgId, collection);
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
    collection: Collection;
}

export const CollectionMenuCell = memo(View);
