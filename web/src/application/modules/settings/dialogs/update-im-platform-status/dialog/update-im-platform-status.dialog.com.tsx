import { useEffect, useState } from "react";

import { Dialog, DialogBody, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ImServiceCommands } from "~/settings/data/commands";
import { ImServiceQueries } from "~/settings/data/queries";

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { ESettingStatus } from "@application/shared/enums";
import { useConditionalModule } from "@application/shared/permissions";

import { Separator } from "@/components/ui";

import { UpdateImPlatformStatusForm } from "../form";
import { useUpdateImPlatformStatusDialogState } from "../hooks";
import type { UpdateImPlatformStatusFormOutput } from "../schemas";

export function UpdateImPlatformStatusDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useUpdateImPlatformStatusDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Settings });

    const { mutate: updateStatus, isPending } = ImServiceCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success("IM platform status updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
        onError: dialogOptions?.onError,
    });

    useEffect(() => {
        if (state.mode === "closed") {
            setHasChanges(false);
            clearDialog();
        }
    }, [clearDialog, state.mode]);

    const detailId = state.mode === "open" ? state.id : "";
    const detailQuery = ImServiceQueries.useFindOneById({ id: detailId }, { enabled: state.mode === "open" });
    const imPlatform = detailQuery.data?.data;

    function onSubmit(values: UpdateImPlatformStatusFormOutput) {
        if (state.mode !== "open" || !imPlatform) return;

        updateStatus({
            id: imPlatform.id,
            payload: {
                updateVer: imPlatform.updateVer,
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
    const initialValues = imPlatform
        ? {
              status: imPlatform.status === ESettingStatus.Disabled ? ESettingStatus.Disabled : ESettingStatus.Active,
              expireAt: imPlatform.expireAt ?? undefined,
              inheritable: Boolean(imPlatform.inheritable),
              default: imPlatform.default ?? false,
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
                    <UpdateImPlatformStatusForm
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
