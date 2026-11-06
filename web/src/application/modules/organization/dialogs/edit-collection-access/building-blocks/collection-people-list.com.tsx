import { memo } from "react";

import { cn } from "@/lib/utils";
import { Badge, type StatusMeaning, statusClassName } from "@components/ui/badge";
import type { CollectionMember } from "~/organization/domain";
import { CollectionPermissionBadge, OrgRoleBadge } from "~/organization/module-shared/components";

interface KeyState {
    label: string;
    meaning: StatusMeaning;
}

/**
 * Where a person stands with the collection's key. Somebody suspended keeps their access and is given nothing; somebody
 * whose app has published no identity cannot be given the key yet.
 */
function getKeyState(member: CollectionMember): KeyState {
    if (!member.served) {
        return { label: "Suspended", meaning: "inactive" };
    }

    if (member.identity === null) {
        return { label: "No encryption key yet", meaning: "attention" };
    }

    if (member.hasKey) {
        return { label: "Has the key", meaning: "success" };
    }

    return { label: "Waiting for the key", meaning: "waiting" };
}

/**
 * Everybody with access to the collection - through their role, their grant or their groups' - with their key's state
 * and the fingerprint the app compares before giving it to them.
 */
function View({ members, className }: Props) {
    if (members.length === 0) {
        return <span className="py-2 text-sm text-muted-foreground">Nobody has access yet.</span>;
    }

    return (
        <ul className={cn("flex flex-col rounded-md border", className)}>
            {members.map(member => {
                const keyState = getKeyState(member);

                return (
                    <li
                        key={member.userId}
                        className="flex flex-col gap-2 border-b px-3 py-2.5 last:border-b-0"
                    >
                        <div className="flex flex-wrap items-center gap-x-3 gap-y-1.5">
                            <span className="flex min-w-0 flex-1 flex-col">
                                <span className="truncate text-sm font-medium">
                                    {member.fullName || member.username}
                                </span>
                                {member.fullName ? (
                                    <span className="truncate text-xs text-muted-foreground">{member.username}</span>
                                ) : null}
                            </span>
                            <div className="flex flex-wrap items-center gap-1.5">
                                <OrgRoleBadge role={member.role} />
                                <CollectionPermissionBadge permission={member.permission} />
                                <Badge className={cn(statusClassName(keyState.meaning))}>{keyState.label}</Badge>
                            </div>
                        </div>

                        {member.identity === null ? (
                            <span className="text-xs text-muted-foreground">
                                No encryption key yet: they sync the shellwing app with this server once.
                            </span>
                        ) : (
                            <div className="flex flex-wrap items-baseline gap-x-2">
                                <span className="text-xs text-muted-foreground">Fingerprint</span>
                                <code className="font-mono text-sm tracking-wider text-foreground select-all [word-spacing:0.5em]">
                                    {member.identity.fingerprint}
                                </code>
                            </div>
                        )}
                    </li>
                );
            })}
        </ul>
    );
}

interface Props {
    members: readonly CollectionMember[];
    className?: string;
}

export const CollectionPeopleList = memo(View);
