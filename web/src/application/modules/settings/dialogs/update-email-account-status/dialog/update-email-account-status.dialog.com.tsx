import { useEffect, useState } from "react";

import { Dialog, DialogBody, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { EmailCommands } from "~/settings/data/commands";
import { EmailQueries } from "~/settings/data/queries";

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { ESettingStatus } from "@application/shared/enums";
import { useConditionalModule } from "@application/shared/permissions";

import { Separator } from "@/components/ui";

import { UpdateEmailAccountStatusForm } from "../form";
import { useUpdateEmailAccountStatusDialogState } from "../hooks";
import type { UpdateEmailAccountStatusFormOutput } from "../schemas";

export function UpdateEmailAccountStatusDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useUpdateEmailAccountStatusDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Settings });

    const { mutate: updateStatus, isPending } = EmailCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success("Email account status updated successfully");
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
    const detailQuery = EmailQueries.useFindOneById({ id: detailId }, { enabled: state.mode === "open" });
    const emailAccount = detailQuery.data?.data;

    function onSubmit(values: UpdateEmailAccountStatusFormOutput) {
        if (state.mode !== "open" || !emailAccount) return;

        updateStatus({
            id: emailAccount.id,
            payload: {
                updateVer: emailAccount.updateVer,
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
    const initialValues = emailAccount
        ? {
              status: emailAccount.status === ESettingStatus.Disabled ? ESettingStatus.Disabled : ESettingStatus.Active,
              expireAt: emailAccount.expireAt ?? undefined,
              inheritable: Boolean(emailAccount.inheritable),
              default: emailAccount.default ?? false,
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
                    <UpdateEmailAccountStatusForm
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
