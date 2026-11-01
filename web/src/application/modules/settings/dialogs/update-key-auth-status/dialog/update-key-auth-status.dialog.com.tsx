import { useEffect, useState } from "react";

import { Dialog, DialogBody, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { KeyAuthCommands } from "~/settings/data/commands";
import { KeyAuthQueries } from "~/settings/data/queries";

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { ESettingStatus } from "@application/shared/enums";
import { useConditionalModule } from "@application/shared/permissions";

import { Separator } from "@/components/ui";

import { UpdateKeyAuthStatusForm } from "../form";
import { useUpdateKeyAuthStatusDialogState } from "../hooks";
import type { UpdateKeyAuthStatusFormOutput } from "../schemas";

export function UpdateKeyAuthStatusDialog() {
    const { state, props: dialogOptions, close: closeDialog, clear: clearDialog } = useUpdateKeyAuthStatusDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Settings });

    const { mutate: updateStatus, isPending } = KeyAuthCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success("Key auth status updated successfully");
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
    const detailQuery = KeyAuthQueries.useFindOneById({ id: detailId }, { enabled: state.mode === "open" });
    const keyAuth = detailQuery.data?.data;

    function onSubmit(values: UpdateKeyAuthStatusFormOutput) {
        if (state.mode !== "open" || !keyAuth) return;

        updateStatus({
            id: keyAuth.id,
            payload: {
                updateVer: keyAuth.updateVer,
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
    const initialValues = keyAuth
        ? {
              status: keyAuth.status === ESettingStatus.Disabled ? ESettingStatus.Disabled : ESettingStatus.Active,
              expireAt: keyAuth.expireAt ?? undefined,
              inheritable: Boolean(keyAuth.inheritable),
              default: keyAuth.default ?? false,
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
                    <UpdateKeyAuthStatusForm
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
