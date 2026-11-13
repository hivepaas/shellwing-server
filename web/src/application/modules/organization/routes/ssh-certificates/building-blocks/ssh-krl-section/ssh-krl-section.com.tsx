import { useState } from "react";

import { Button } from "@components/ui/button";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { format } from "date-fns";
import { KeyRound, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { SshKrlCommands } from "~/organization/data/commands";
import { PublicKeyBlock } from "~/organization/module-shared/components";
import {
    SSHD_REVOKED_KEYS_LINE,
    SSH_FETCH_DEPLOYMENT_PATH,
    SSH_KRL_PATH,
    sshCaKeysApiPath,
    sshKrlApiPath,
} from "~/organization/module-shared/constants";

import { ContentBlock, InfoBlock, PopConfirm } from "@application/shared/components";

interface Props {
    orgId: string;
    tokenCreatedAt: Date | null;
}

export function SshKrlSection({ orgId, tokenCreatedAt }: Props) {
    // Shown once, right after it is made: only its hash is kept, so there is nothing to come back for.
    const [token, setToken] = useState<string | null>(null);

    const { mutate: createToken, isPending: isCreating } = SshKrlCommands.useCreateOne({
        onSuccess: response => {
            setToken(response.data.token);
            toast.success("Token made. Copy it now: it is not shown again.");
        },
    });
    const { mutate: deleteToken, isPending: isDeleting } = SshKrlCommands.useDeleteOne({
        onSuccess: () => {
            setToken(null);
            toast.success("Token taken away. Your servers can no longer fetch the list.");
        },
    });

    return (
        <ContentBlock label="Revocation list">
            <div className="flex flex-col gap-6">
                <p className="text-sm text-muted-foreground">
                    A certificate cannot be taken back once signed, so taking somebody&apos;s accounts away puts the
                    certificates they hold on this list. Your servers read it and refuse those certificates for the rest
                    of their lifetime; until a server has fetched the list again, what was revoked still works there.
                </p>

                <div className="flex flex-col gap-2">
                    <span className="text-sm font-medium">
                        Add this to <code className="font-mono text-xs">/etc/ssh/sshd_config</code>
                    </span>
                    <PublicKeyBlock
                        value={SSHD_REVOKED_KEYS_LINE}
                        label="Copy the line"
                    />
                </div>

                <div className="flex flex-col gap-3">
                    <div className="flex flex-col gap-1">
                        <span className="text-sm font-medium">A token for your servers</span>
                        <span className="text-sm text-muted-foreground">
                            A server has no account here, so it fetches with a token of its own. It fetches this
                            organization&apos;s list and does nothing else - unlike a person&apos;s session, which can
                            do everything that person can.
                        </span>
                    </div>

                    {tokenCreatedAt === null && (
                        <div className={cn(dashedBorderBox, "flex flex-col items-start gap-3")}>
                            <span>
                                Your servers have no token, so nothing fetches the list except a member signed in here.
                            </span>
                            <Button
                                type="button"
                                isLoading={isCreating}
                                onClick={() => {
                                    createToken({ orgId });
                                }}
                            >
                                {!isCreating && <KeyRound className="size-4" />}
                                Make a token
                            </Button>
                        </div>
                    )}

                    {tokenCreatedAt !== null && (
                        <div className="flex flex-wrap items-center justify-between gap-3">
                            <span className="text-sm text-muted-foreground">
                                Made {format(tokenCreatedAt, "yyyy-MM-dd HH:mm")}
                            </span>
                            <div className="flex flex-wrap items-center gap-2">
                                <Button
                                    type="button"
                                    variant="outline"
                                    isLoading={isCreating}
                                    onClick={() => {
                                        createToken({ orgId });
                                    }}
                                >
                                    {!isCreating && <KeyRound className="size-4" />}
                                    Replace
                                </Button>
                                <PopConfirm
                                    title="Take the token away?"
                                    description="Your servers stop being able to fetch the list. The list each has in place keeps working."
                                    onConfirm={() => {
                                        deleteToken({ orgId });
                                    }}
                                >
                                    <Button
                                        type="button"
                                        variant="outline"
                                        isLoading={isDeleting}
                                    >
                                        {!isDeleting && <Trash2Icon className="size-4" />}
                                        Take away
                                    </Button>
                                </PopConfirm>
                            </div>
                        </div>
                    )}

                    {token !== null && (
                        <div className="flex flex-col gap-2">
                            <InfoBlock
                                title="Copy it now"
                                description="It is not shown again - only its hash is kept here."
                            />
                            <PublicKeyBlock
                                value={token}
                                label="Copy the token"
                            />
                        </div>
                    )}
                </div>

                <div className="flex flex-col gap-1 text-sm text-muted-foreground">
                    <span>
                        Each server then fetches <code className="font-mono text-xs">{sshKrlApiPath(orgId)}</code> with
                        that token in the <code className="font-mono text-xs">X-Shellwing-KRL-Token</code> header, and
                        writes it to <code className="font-mono text-xs">{SSH_KRL_PATH}</code>. The same token fetches{" "}
                        <code className="font-mono text-xs">{sshCaKeysApiPath(orgId)}</code>, the authority keys the
                        server trusts, which is what lets an authority be rotated without editing every server. A script
                        and a systemd timer that do both are in{" "}
                        <code className="font-mono text-xs">{SSH_FETCH_DEPLOYMENT_PATH}</code>; sshd reads both files on
                        each login, so nothing needs reloading.
                    </span>
                    <span>
                        Nothing is pushed to your servers - this server cannot reach them and does not know they exist -
                        so how often they fetch is how long a revocation takes to arrive.
                    </span>
                </div>
            </div>
        </ContentBlock>
    );
}
