import { useEffect, useState } from "react";

import { Dialog, DialogBody, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { CloudStorageCommands } from "~/settings/data/commands";
import { CloudStorageQueries } from "~/settings/data/queries";

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { ESettingStatus } from "@application/shared/enums";
import { useConditionalModule } from "@application/shared/permissions";

import { Separator } from "@/components/ui";

import { UpdateCloudStorageStatusForm } from "../form";
import { useUpdateCloudStorageStatusDialogState } from "../hooks";
import type { UpdateCloudStorageStatusFormOutput } from "../schemas";

export function UpdateCloudStorageStatusDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useUpdateCloudStorageStatusDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Settings });

    const { mutate: updateStatus, isPending } = CloudStorageCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success("Cloud storage status updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });

    useEffect(() => {
        if (state.mode === "closed") {
            setHasChanges(false);
            clearDialog();
        }
    }, [clearDialog, state.mode]);

    const detailId = state.mode === "open" ? state.id : "";
    const detailQuery = CloudStorageQueries.useFindOneById({ id: detailId }, { enabled: state.mode === "open" });
    const cloudStorage = detailQuery.data?.data;

    function onSubmit(values: UpdateCloudStorageStatusFormOutput) {
        if (state.mode !== "open" || !cloudStorage) return;

        updateStatus({
            id: cloudStorage.id,
            payload: {
                updateVer: cloudStorage.updateVer,
                status: values.status,
                expireAt: values.expireAt ?? null,
                inheritable: values.inheritable,
                default: values.default,
            },
        });
    }

    function handleClose() {
        if (isPending) return;
        if (canWrite && hasChanges && !window.confirm("Are you sure you want to close without saving changes?")) return;
        closeDialog();
        dialogOptions?.onClose?.();
    }

    const open = state.mode !== "closed";
    const initialValues = cloudStorage
        ? {
              status: cloudStorage.status === ESettingStatus.Disabled ? ESettingStatus.Disabled : ESettingStatus.Active,
              expireAt: cloudStorage.expireAt ?? undefined,
              inheritable: Boolean(cloudStorage.inheritable),
              default: cloudStorage.default ?? false,
          }
        : undefined;
    const isDetailLoading = state.mode === "open" && detailQuery.isFetching;

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogFixedContent className="sm:max-w-[560px]">
                <DialogHeader>
                    <DialogTitle>Change status</DialogTitle>
                </DialogHeader>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>
                {isDetailLoading && (
                    <DialogBody>
                        <AppLoader />
                    </DialogBody>
                )}
                {state.mode === "open" && !isDetailLoading && initialValues && (
                    <UpdateCloudStorageStatusForm
                        isPending={isPending}
                        onSubmit={onSubmit}
                        onHasChanges={setHasChanges}
                        initialValues={initialValues}
                        showAvailableInProjects
                        readOnly={!canWrite}
                        onClose={handleClose}
                    />
                )}
            </DialogFixedContent>
        </Dialog>
    );
}
