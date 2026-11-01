import { useEffect, useState } from "react";

import { Dialog, DialogBody, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { NotificationCommands } from "~/settings/data/commands";
import { NotificationQueries } from "~/settings/data/queries";

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { ESettingStatus } from "@application/shared/enums";
import { useConditionalModule } from "@application/shared/permissions";

import { Separator } from "@/components/ui";

import { UpdateNotificationTargetStatusForm } from "../form";
import { useUpdateNotificationTargetStatusDialogState } from "../hooks";
import type { UpdateNotificationTargetStatusFormOutput } from "../schemas";

export function UpdateNotificationTargetStatusDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useUpdateNotificationTargetStatusDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Settings });

    const { mutate: updateStatus, isPending } = NotificationCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success("Notification target status updated successfully");
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
    const detailQuery = NotificationQueries.useFindOneById({ id: detailId }, { enabled: state.mode === "open" });
    const notificationTarget = detailQuery.data?.data;

    function onSubmit(values: UpdateNotificationTargetStatusFormOutput) {
        if (state.mode !== "open" || !notificationTarget) return;

        updateStatus({
            id: notificationTarget.id,
            payload: {
                updateVer: notificationTarget.updateVer,
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
    const initialValues = notificationTarget
        ? {
              status:
                  notificationTarget.status === ESettingStatus.Disabled
                      ? ESettingStatus.Disabled
                      : ESettingStatus.Active,
              expireAt: notificationTarget.expireAt ?? undefined,
              inheritable: Boolean(notificationTarget.inheritable),
              default: notificationTarget.default ?? false,
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
                    <UpdateNotificationTargetStatusForm
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
