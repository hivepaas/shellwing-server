import { Button } from "@components/ui/button";
import { Skeleton } from "@components/ui/skeleton";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { format } from "date-fns";
import { Copy } from "lucide-react";
import { toast } from "sonner";
import { UserIdentityQueries } from "~/user-management/data/queries";

import { InfoBlock } from "@application/shared/components";

import { Separator } from "@/components/ui/separator";

/**
 * The fingerprint of the identity the person's shellwing app published - colleagues compare it before sharing a
 * collection's key with them.
 */
export function EncryptionKey() {
    const { data, isLoading, error } = UserIdentityQueries.useFindMine();

    const identity = data?.data ?? null;

    function handleCopy(fingerprint: string) {
        void navigator.clipboard
            .writeText(fingerprint)
            .then(() => {
                toast.success("Fingerprint copied to clipboard");
            })
            .catch(() => {
                toast.error("Failed to copy");
            });
    }

    return (
        <>
            <Separator className="opacity-50" />
            <InfoBlock
                titleWidth={220}
                title="Encryption key"
            >
                {isLoading && <Skeleton className="h-10 w-full max-w-[420px]" />}

                {!isLoading && data === undefined && (
                    <div className={cn(dashedBorderBox)}>
                        <span className="font-semibold text-destructive">Failed to get your encryption key:</span>{" "}
                        {error?.message ?? "Unknown error"}
                    </div>
                )}

                {data !== undefined && identity === null && (
                    <p className="text-sm text-muted-foreground">
                        Your fingerprint appears here once you connect the shellwing app to this server (Settings › Sync
                        in the app).
                    </p>
                )}

                {identity !== null && (
                    <div className="flex flex-col gap-2">
                        <div className="flex w-fit max-w-full items-center gap-2 rounded-md border bg-muted px-3 py-2">
                            <code className="min-w-0 font-mono text-lg tracking-wider text-foreground select-all [word-spacing:0.5em]">
                                {identity.fingerprint}
                            </code>
                            <Button
                                type="button"
                                variant="ghost"
                                size="icon"
                                className="size-7 shrink-0"
                                aria-label="Copy the fingerprint"
                                onClick={() => {
                                    handleCopy(identity.fingerprint);
                                }}
                            >
                                <Copy className="size-4 text-muted-foreground" />
                            </Button>
                        </div>
                        <p className="text-sm text-muted-foreground">
                            Colleagues compare this with what their app shows before they share a collection with you.
                        </p>
                        <p className="text-sm text-muted-foreground">
                            Last published {format(identity.updatedAt, "yyyy-MM-dd HH:mm:ss")}
                        </p>
                    </div>
                )}
            </InfoBlock>
        </>
    );
}
