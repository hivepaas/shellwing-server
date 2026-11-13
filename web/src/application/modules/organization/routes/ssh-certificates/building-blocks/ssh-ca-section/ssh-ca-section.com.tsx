import { Badge, statusClassName } from "@components/ui/badge";
import { Button } from "@components/ui/button";
import { Skeleton } from "@components/ui/skeleton";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { format } from "date-fns";
import { Plus, RefreshCw, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { SshCaCommands } from "~/organization/data/commands";
import { SshCaQueries } from "~/organization/data/queries";
import { useReplaceSshCaDialog } from "~/organization/dialogs";
import { ESshCaStatus, type SshCa } from "~/organization/domain";
import { PublicKeyBlock } from "~/organization/module-shared/components";
import {
    SSHD_REVOKED_KEYS_LINE,
    SSHD_TRUSTED_USER_CA_KEYS_LINE,
    SSH_CA_PUBLIC_KEY_PATH,
    SSH_KRL_PATH,
    sshKrlApiPath,
} from "~/organization/module-shared/constants";

import { ContentBlock, PopConfirm } from "@application/shared/components";

const RELOAD_SSHD = "sudo systemctl reload sshd";
const RELOAD_SSH = "sudo systemctl reload ssh";

function formatDate(date: Date) {
    return format(date, "yyyy-MM-dd HH:mm");
}

function TrustHowTo({ orgId }: { orgId: string }) {
    return (
        <div className="flex flex-col gap-3 text-sm">
            <span className="font-medium">Have each server trust it</span>
            <ol className="ml-5 flex list-decimal flex-col gap-3">
                <li>
                    Save the public key above on the server as{" "}
                    <code className="font-mono text-xs">{SSH_CA_PUBLIC_KEY_PATH}</code>.
                </li>
                <li className="flex flex-col gap-2">
                    <span>
                        Add this line to <code className="font-mono text-xs">/etc/ssh/sshd_config</code>:
                    </span>
                    <PublicKeyBlock
                        value={SSHD_TRUSTED_USER_CA_KEYS_LINE}
                        label="Copy the line"
                    />
                </li>
                <li className="flex flex-col gap-2">
                    <span>Add this line as well, so the server refuses a certificate that has been revoked:</span>
                    <PublicKeyBlock
                        value={SSHD_REVOKED_KEYS_LINE}
                        label="Copy the line"
                    />
                    <span className="text-muted-foreground">
                        Then keep <code className="font-mono text-xs">{SSH_KRL_PATH}</code> up to date from{" "}
                        <code className="font-mono text-xs">{sshKrlApiPath(orgId)}</code>, the way the rest of the
                        server&apos;s configuration is. Write it to a temporary name and move it into place: a server
                        told to read the file and unable to refuses every certificate.
                    </span>
                </li>
                <li>
                    Reload sshd: <code className="font-mono text-xs">{RELOAD_SSHD}</code>, or{" "}
                    <code className="font-mono text-xs">{RELOAD_SSH}</code> on Debian and Ubuntu.
                </li>
            </ol>
            <p className="text-muted-foreground">
                A member then logs in with a certificate this server signs for them, as one of the accounts granted to
                them below. sshd logs the person and the device of each login. Taking an account away revokes the
                certificates they hold, which reaches this server once it has fetched the list again.
            </p>
        </div>
    );
}

function ActiveCa({ orgId, ca, onReplace }: { orgId: string; ca: SshCa; onReplace: () => void }) {
    return (
        <div className="flex flex-col gap-4">
            <div className="flex flex-wrap items-center justify-between gap-3">
                <div className="flex items-center gap-2 text-sm">
                    <Badge className={cn(statusClassName("success"))}>Active</Badge>
                    <span className="text-muted-foreground">Created {formatDate(ca.createdAt)}</span>
                </div>
                <Button
                    type="button"
                    variant="destructiveOutline"
                    onClick={onReplace}
                >
                    <RefreshCw className="size-4" />
                    Replace
                </Button>
            </div>
            <div className="flex flex-col gap-2">
                <span className="text-sm font-medium">Public key</span>
                <PublicKeyBlock
                    value={ca.publicKey}
                    label="Copy the public key"
                />
            </div>
            <TrustHowTo orgId={orgId} />
        </div>
    );
}

function RetiredCa({ orgId, ca }: { orgId: string; ca: SshCa }) {
    const { mutate: deleteCa, isPending } = SshCaCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Certificate authority deleted");
        },
    });

    return (
        <li className="flex flex-col gap-2 rounded-lg border p-3">
            <div className="flex flex-wrap items-center justify-between gap-2">
                <div className="flex flex-wrap items-center gap-2 text-sm">
                    <Badge className={cn(statusClassName("inactive"))}>Retired</Badge>
                    <span className="text-muted-foreground">
                        Created {formatDate(ca.createdAt)}
                        {ca.retiredAt ? ` · Retired ${formatDate(ca.retiredAt)}` : null}
                    </span>
                </div>
                <PopConfirm
                    title="Delete this certificate authority?"
                    description="Its public key is no longer listed here. Remove it from your servers too, once the certificates it signed have ended."
                    confirmText="Delete"
                    variant="destructive"
                    align="end"
                    onConfirm={() => {
                        deleteCa({ orgId, caId: ca.id });
                    }}
                >
                    <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        isLoading={isPending}
                    >
                        {!isPending && <Trash2Icon className="size-4" />}
                        Delete
                    </Button>
                </PopConfirm>
            </div>
            <PublicKeyBlock
                value={ca.publicKey}
                label="Copy the public key"
            />
        </li>
    );
}

export function SshCaSection({ orgId }: Props) {
    const replaceDialog = useReplaceSshCaDialog();
    const { data, isLoading, error } = SshCaQueries.useFindMany({ orgId });
    const { mutate: createCa, isPending: isCreating } = SshCaCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Certificate authority created. Add its public key to your servers.");
        },
    });

    const cas = data?.data ?? [];
    const active = cas.find(ca => ca.status === ESshCaStatus.Active) ?? null;
    const retired = cas.filter(ca => ca.status !== ESshCaStatus.Active);

    return (
        <ContentBlock label="Certificate authority">
            <div className="flex flex-col gap-6">
                <p className="text-sm text-muted-foreground">
                    This server signs short-lived SSH certificates, so that nobody shares a private key and removing a
                    person means they get no new ones. Your servers accept them once they trust the organization&apos;s
                    certificate authority.
                </p>

                {isLoading && (
                    <div className="flex flex-col gap-3">
                        <Skeleton className="h-6 w-40" />
                        <Skeleton className="h-16 w-full" />
                    </div>
                )}

                {!isLoading && data === undefined && (
                    <div className={cn(dashedBorderBox)}>
                        <span className="font-semibold text-destructive">
                            Failed to get the certificate authorities:
                        </span>{" "}
                        {error?.message ?? "Unknown error"}
                    </div>
                )}

                {data !== undefined && active === null && (
                    <div className={cn(dashedBorderBox, "flex flex-col items-start gap-3")}>
                        <span>
                            There is no certificate authority yet, so nobody can get a certificate. Make one, then have
                            each server trust its public key.
                        </span>
                        <Button
                            type="button"
                            isLoading={isCreating}
                            onClick={() => {
                                createCa({ orgId });
                            }}
                        >
                            {!isCreating && <Plus className="size-4" />}
                            Create certificate authority
                        </Button>
                    </div>
                )}

                {active !== null && (
                    <ActiveCa
                        orgId={orgId}
                        ca={active}
                        onReplace={() => {
                            replaceDialog.actions.open(orgId);
                        }}
                    />
                )}

                {retired.length > 0 && (
                    <div className="flex flex-col gap-3">
                        <div className="flex flex-col gap-1">
                            <span className="text-sm font-medium">Retired</span>
                            <span className="text-sm text-muted-foreground">
                                They sign no more, and their private keys are gone. Keep one on a server only until the
                                certificates it signed have ended, then remove it there and delete it here.
                            </span>
                        </div>
                        <ul className="flex flex-col gap-3">
                            {retired.map(ca => (
                                <RetiredCa
                                    key={ca.id}
                                    orgId={orgId}
                                    ca={ca}
                                />
                            ))}
                        </ul>
                    </div>
                )}
            </div>
        </ContentBlock>
    );
}

interface Props {
    orgId: string;
}
