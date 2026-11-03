import { memo } from "react";

import { Button } from "@components/ui/button";
import { Copy } from "lucide-react";
import { toast } from "sonner";

/**
 * A value to be copied as it is, in monospace, with a button that copies it.
 */
function View({ value, label = "Copy" }: Props) {
    function handleCopy() {
        void navigator.clipboard
            .writeText(value)
            .then(() => {
                toast.success("Copied to clipboard");
            })
            .catch(() => {
                toast.error("Failed to copy");
            });
    }

    return (
        <div className="flex w-fit max-w-full items-center gap-2 rounded-md border bg-muted px-3 py-1.5">
            <code className="min-w-0 break-all font-mono text-sm text-foreground select-all">{value}</code>
            <Button
                type="button"
                variant="ghost"
                size="icon"
                className="size-7 shrink-0"
                aria-label={label}
                onClick={handleCopy}
            >
                <Copy className="size-4 text-muted-foreground" />
            </Button>
        </div>
    );
}

interface Props {
    value: string;
    label?: string;
}

export const CopyableValue = memo(View);
