import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { CloudStorageCommands } from "~/settings/data/commands";
import { useUpdateCloudStorageStatusDialog } from "~/settings/dialogs/update-cloud-storage-status";
import type { SettingCloudStorage } from "~/settings/domain";
import { SettingsMenuButton, SettingsPopConfirmButton } from "~/settings/module-shared/components";

import { CopyIdMenuButton } from "@application/shared/components";

function View({ cloudStorage }: Props) {
    const [open, setOpen] = useState(false);
    const updateStatusDialog = useUpdateCloudStorageStatusDialog();
    const { mutate: deleteCloudStorage, isPending: isDeleting } = CloudStorageCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Cloud storage deleted successfully");
            setOpen(false);
        },
    });

    function handleDelete() {
        deleteCloudStorage({ id: cloudStorage.id });
    }

    function handleChangeStatus() {
        updateStatusDialog.actions.open(cloudStorage.id);
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
                        id={cloudStorage.id}
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
                        title="Delete cloud storage"
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
    cloudStorage: SettingCloudStorage;
}

export const CloudStorageMenuCell = memo(View);
