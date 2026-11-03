import { memo } from "react";

import { Button } from "@components/ui/button";
import { cn } from "@lib/utils";
import { Copy } from "lucide-react";
import { toast } from "sonner";

/**
 * A line to be copied as it is - a public key, a line of configuration - in monospace, with a button that copies it.
 */
function View({ value, label = "Copy", className }: Props) {
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
        <div className={cn("flex items-start gap-2 rounded-md border bg-muted px-3 py-2", className)}>
            <code className="min-w-0 flex-1 break-all font-mono text-xs leading-5 text-foreground select-all">
                {value}
            </code>
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
    className?: string;
}

export const PublicKeyBlock = memo(View);
