import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ImServiceCommands } from "~/settings/data/commands";
import { useUpdateImPlatformStatusDialog } from "~/settings/dialogs/update-im-platform-status";
import type { SettingImService } from "~/settings/domain";
import { SettingsMenuButton, SettingsPopConfirmButton } from "~/settings/module-shared/components";

import { CopyIdMenuButton } from "@application/shared/components";

function View({ imPlatform }: Props) {
    const [open, setOpen] = useState(false);
    const updateStatusDialog = useUpdateImPlatformStatusDialog();
    const { mutate: deleteImPlatform, isPending: isDeleting } = ImServiceCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("IM platform deleted successfully");
            setOpen(false);
        },
    });

    function handleDelete() {
        deleteImPlatform({ id: imPlatform.id });
    }

    function handleChangeStatus() {
        updateStatusDialog.actions.open(imPlatform.id);
        setOpen(false);
    }

    return (
        <DropdownMenu
            open={open}
            onOpenChange={setOpen}
        >
            <DropdownMenuTrigger asChild>
                <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8"
                >
                    <MoreVertical className="size-4" />
                    <span className="sr-only">Actions menu</span>
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
                <div className="flex flex-col gap-0">
                    <CopyIdMenuButton
                        id={imPlatform.id}
                        onCopied={() => {
                            setOpen(false);
                        }}
                    />
                    <SettingsMenuButton
                        action="write"
                        onClick={handleChangeStatus}
                    >
                        <SlidersHorizontal className="mr-2 size-4" />
                        Change Status
                    </SettingsMenuButton>
                    <SettingsPopConfirmButton
                        action="delete"
                        title="Delete IM platform"
                        confirmText="Delete"
                        cancelText="Cancel"
                        description="Confirm deletion of this item?"
                        onConfirm={handleDelete}
                        isLoading={isDeleting}
                    >
                        <Trash2Icon className="mr-2 size-4" />
                        Delete
                    </SettingsPopConfirmButton>
                </div>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

interface Props {
    imPlatform: SettingImService;
}

export const ImPlatformMenuCell = memo(View);
