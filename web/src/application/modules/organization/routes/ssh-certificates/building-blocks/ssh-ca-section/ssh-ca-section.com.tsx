import { Badge, statusClassName } from "@components/ui/badge";
import { Button } from "@components/ui/button";
import { Skeleton } from "@components/ui/skeleton";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { format } from "date-fns";
import { ArrowUpCircle, CircleSlash, Plus, RefreshCw, Trash2Icon } from "lucide-react";
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
    SSH_FETCH_DEPLOYMENT_PATH,
    SSH_KRL_PATH,
    SSH_MAX_TRUSTED_CAS,
    sshCaKeysApiPath,
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
            <span className="font-medium">Have each server trust this organization</span>
            <ol className="ml-5 flex list-decimal flex-col gap-3">
                <li className="flex flex-col gap-2">
                    <span>
                        Add both lines to <code className="font-mono text-xs">/etc/ssh/sshd_config</code>:
                    </span>
                    <PublicKeyBlock
                        value={`${SSHD_TRUSTED_USER_CA_KEYS_LINE}\n${SSHD_REVOKED_KEYS_LINE}`}
                        label="Copy both lines"
                    />
                    <span className="text-muted-foreground">
                        The first says whose certificates the server accepts, the second which ones it refuses before
                        they expire. They are useless apart: with the keys and no list a server takes certificates that
                        were revoked, and with the list and no keys it takes none at all.
                    </span>
                </li>
                <li className="flex flex-col gap-2">
                    <span>Keep both files up to date from this server:</span>
                    <PublicKeyBlock
                        value={`${sshCaKeysApiPath(orgId)}\n${sshKrlApiPath(orgId)}`}
                        label="Copy both paths"
                    />
                    <span className="text-muted-foreground">
                        Into <code className="font-mono text-xs">{SSH_CA_PUBLIC_KEY_PATH}</code> and{" "}
                        <code className="font-mono text-xs">{SSH_KRL_PATH}</code>, with the token below. A script and a
                        systemd timer that do it are in{" "}
                        <code className="font-mono text-xs">{SSH_FETCH_DEPLOYMENT_PATH}</code>. Fetching the keys rather
                        than copying them once is what lets an authority be rotated without locking anybody out.
                    </span>
                </li>
                <li>
                    Reload sshd: <code className="font-mono text-xs">{RELOAD_SSHD}</code>, or{" "}
                    <code className="font-mono text-xs">{RELOAD_SSH}</code> on Debian and Ubuntu. After that nothing
                    needs reloading: sshd reads both files on every login.
                </li>
            </ol>
            <p className="text-muted-foreground">
                A member then logs in with a certificate this server signs for them, as one of the accounts granted to
                them below. sshd logs the person and the device of each login. Taking an account away revokes the
                certificates they hold, which reaches a server once it has fetched the list again.
            </p>
        </div>
    );
}

function ActiveCa({
    orgId,
    ca,
    canPublish,
    isPublishing,
    onPublish,
    onReplace,
}: {
    orgId: string;
    ca: SshCa;
    canPublish: boolean;
    isPublishing: boolean;
    onPublish: () => void;
    onReplace: () => void;
}) {
    return (
        <div className="flex flex-col gap-4">
            <div className="flex flex-wrap items-center justify-between gap-3">
                <div className="flex items-center gap-2 text-sm">
                    <Badge className={cn(statusClassName("success"))}>Signing</Badge>
                    <span className="text-muted-foreground">Created {formatDate(ca.createdAt)}</span>
                </div>
                <div className="flex flex-wrap items-center gap-2">
                    {canPublish && (
                        <Button
                            type="button"
                            variant="outline"
                            isLoading={isPublishing}
                            onClick={onPublish}
                        >
                            {!isPublishing && <Plus className="size-4" />}
                            Publish a new one
                        </Button>
                    )}
                    <Button
                        type="button"
                        variant="destructiveOutline"
                        onClick={onReplace}
                    >
                        <RefreshCw className="size-4" />
                        Replace now
                    </Button>
                </div>
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

function PublishedCa({ orgId, ca }: { orgId: string; ca: SshCa }) {
    const { mutate: setStatus, isPending } = SshCaCommands.useSetStatus({
        onSuccess: response => {
            toast.success(
                response.data.status === ESshCaStatus.Active
                    ? "That authority signs from now on. The one it replaced stays published, so certificates already held keep working."
                    : "Certificate authority retired. Its certificates are now refused.",
            );
        },
    });

    return (
        <li className="flex flex-col gap-2 rounded-lg border p-3">
            <div className="flex flex-wrap items-center justify-between gap-2">
                <div className="flex flex-wrap items-center gap-2 text-sm">
                    <Badge className={cn(statusClassName("waiting"))}>Published</Badge>
                    <span className="text-muted-foreground">Created {formatDate(ca.createdAt)}</span>
                </div>
                <div className="flex flex-wrap items-center gap-2">
                    <PopConfirm
                        title="Sign with this authority from now on?"
                        description="Every server has to have its key first, or anybody whose server does not will be refused. The authority it replaces stays published, so certificates already held keep working until they expire."
                        confirmText="Sign with it"
                        align="end"
                        onConfirm={() => {
                            setStatus({ orgId, caId: ca.id, status: ESshCaStatus.Active });
                        }}
                    >
                        <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            isLoading={isPending}
                        >
                            {!isPending && <ArrowUpCircle className="size-4" />}
                            Sign with it
                        </Button>
                    </PopConfirm>
                    <PopConfirm
                        title="Retire this certificate authority?"
                        description="Your servers stop trusting it and every certificate it signed is refused. Do this once nothing it signed is still valid, which is at most one certificate lifetime after it stopped signing."
                        confirmText="Retire"
                        variant="destructive"
                        align="end"
                        onConfirm={() => {
                            setStatus({ orgId, caId: ca.id, status: ESshCaStatus.Retired });
                        }}
                    >
                        <Button
                            type="button"
                            variant="ghost"
                            size="sm"
                            isLoading={isPending}
                        >
                            {!isPending && <CircleSlash className="size-4" />}
                            Retire
                        </Button>
                    </PopConfirm>
                </div>
            </div>
            <PublicKeyBlock
                value={ca.publicKey}
                label="Copy the public key"
            />
        </li>
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
    const published = cas.filter(ca => ca.status === ESshCaStatus.Published);
    const retired = cas.filter(ca => ca.status === ESshCaStatus.Retired);
    // The bound counts every authority the servers trust, which is the one signing and the published ones.
    const canPublish = active !== null && 1 + published.length < SSH_MAX_TRUSTED_CAS;

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
                            each server trust this organization.
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
                        canPublish={canPublish}
                        isPublishing={isCreating}
                        onPublish={() => {
                            createCa({ orgId });
                        }}
                        onReplace={() => {
                            replaceDialog.actions.open(orgId);
                        }}
                    />
                )}

                {published.length > 0 && (
                    <div className="flex flex-col gap-3">
                        <div className="flex flex-col gap-1">
                            <span className="text-sm font-medium">Published</span>
                            <span className="text-sm text-muted-foreground">
                                Your servers trust these, but nothing is signed with them. That is how an authority is
                                rotated: the servers pick up the new key first, then it starts signing, and the one it
                                replaces stays here until the certificates it signed have expired.
                            </span>
                        </div>
                        <ul className="flex flex-col gap-3">
                            {published.map(ca => (
                                <PublishedCa
                                    key={ca.id}
                                    orgId={orgId}
                                    ca={ca}
                                />
                            ))}
                        </ul>
                    </div>
                )}

                {retired.length > 0 && (
                    <div className="flex flex-col gap-3">
                        <div className="flex flex-col gap-1">
                            <span className="text-sm font-medium">Retired</span>
                            <span className="text-sm text-muted-foreground">
                                Not trusted and not signing: their private keys are gone and what they signed is
                                refused. Listed so you can see what was removed; delete one here once no server has it
                                any more.
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
