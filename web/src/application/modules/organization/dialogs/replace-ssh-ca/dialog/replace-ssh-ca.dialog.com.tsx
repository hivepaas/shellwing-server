import { Button } from "@components/ui/button";
import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogDescription,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { toast } from "sonner";
import { SshCaCommands } from "~/organization/data/commands";

import { useReplaceSshCaDialogState } from "../hooks";

export function ReplaceSshCaDialog() {
    const { state, props: dialogProps, ...actions } = useReplaceSshCaDialogState();
    const { mutate: createCa, isPending, reset } = SshCaCommands.useCreateOne();

    const open = state.mode !== "closed";

    function handleClose() {
        if (isPending) {
            return;
        }

        reset();
        actions.close();
        dialogProps?.onClose?.();
    }

    function handleReplace() {
        if (state.mode !== "open") {
            return;
        }

        createCa(
            { orgId: state.orgId },
            {
                onSuccess: () => {
                    toast.success("A new certificate authority is active. Add its public key to your servers.");
                    actions.close();
                    dialogProps?.onSuccess?.();
                },
            },
        );
    }

    return (
        <Dialog
            open={open}
            modal
            onOpenChange={handleClose}
        >
            <DialogFixedContent className="sm:max-w-[560px]">
                <DialogHeader>
                    <DialogTitle>Replace the certificate authority</DialogTitle>
                    <DialogDescription>Make a new one, and retire the one in use</DialogDescription>
                </DialogHeader>

                <DialogBody className="flex flex-col gap-4 text-sm">
                    <p>
                        A new certificate authority signs every certificate from now on. The current one is retired at
                        once and <strong>its private key is destroyed</strong>: it can never sign again, and this cannot
                        be undone.
                    </p>
                    <p>
                        Servers that trust only the current one stop accepting new certificates. Add the new public key
                        to each of them beside the old one straight away, and remove the old one when its last
                        certificates have ended.
                    </p>
                    <div className={cn(dashedBorderBox)}>
                        If the private key may have leaked, remove the old public key from your servers first.
                    </div>
                </DialogBody>

                <DialogActionFooter className="flex items-center justify-end gap-3">
                    <Button
                        type="button"
                        variant="outline"
                        className="min-w-[100px]"
                        disabled={isPending}
                        onClick={handleClose}
                    >
                        Cancel
                    </Button>
                    <Button
                        type="button"
                        variant="destructive"
                        isLoading={isPending}
                        onClick={handleReplace}
                    >
                        Replace
                    </Button>
                </DialogActionFooter>
            </DialogFixedContent>
        </Dialog>
    );
}
