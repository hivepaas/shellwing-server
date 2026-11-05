import React, { useMemo, useState } from "react";

import { Checkbox } from "@components/ui/checkbox";
import { DialogBody } from "@components/ui/dialog";
import { Input } from "@components/ui/input";
import { SearchIcon } from "lucide-react";
import { EOrgMemberStatus, type OrgMember } from "~/organization/domain";

/**
 * The organization's members, each with a box ticked when they are to be in the group. The search narrows the list, not
 * who is ticked.
 */
export function EditOrgGroupMembersForm({ members, defaultUserIds, onSubmit, onHasChanges, footer }: Props) {
    const [selected, setSelected] = useState<ReadonlySet<string>>(() => new Set(defaultUserIds));
    const [search, setSearch] = useState("");

    const isDirty = useMemo(
        () => selected.size !== defaultUserIds.length || defaultUserIds.some(userId => !selected.has(userId)),
        [selected, defaultUserIds],
    );

    React.useEffect(() => {
        onHasChanges?.(isDirty);
    }, [isDirty, onHasChanges]);

    const shown = useMemo(() => {
        const term = search.trim().toLowerCase();

        if (!term) {
            return members;
        }

        return members.filter(
            member => member.username.toLowerCase().includes(term) || member.fullName.toLowerCase().includes(term),
        );
    }, [members, search]);

    function toggle(userId: string, checked: boolean) {
        setSelected(previous => {
            const next = new Set(previous);

            if (checked) {
                next.add(userId);
            } else {
                next.delete(userId);
            }

            return next;
        });
    }

    return (
        <form
            onSubmit={event => {
                event.preventDefault();
                // Who is listed, in the list's order: anybody ticked who is not listed any more is left out.
                onSubmit(members.filter(member => selected.has(member.userId)).map(member => member.userId));
            }}
            className="min-h-0 flex flex-1 flex-col"
        >
            <DialogBody className="flex flex-col gap-3">
                <div className="relative">
                    <SearchIcon className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
                    <Input
                        type="search"
                        value={search}
                        onChange={event => {
                            setSearch(event.target.value);
                        }}
                        placeholder="Search members"
                        aria-label="Search members"
                        className="pl-9"
                    />
                </div>

                <span className="text-sm text-muted-foreground">
                    {selected.size === 1 ? "1 member" : `${selected.size} members`} selected
                </span>

                {shown.length === 0 ? (
                    <span className="py-4 text-center text-sm text-muted-foreground">
                        {members.length === 0 ? "The organization has no members to add." : "No member matches."}
                    </span>
                ) : (
                    <ul className="flex max-h-[320px] flex-col overflow-y-auto rounded-md border">
                        {shown.map(member => {
                            const id = `org-group-member-${member.userId}`;

                            return (
                                <li
                                    key={member.userId}
                                    className="border-b last:border-b-0"
                                >
                                    <label
                                        htmlFor={id}
                                        className="flex cursor-pointer items-center gap-3 px-3 py-2 hover:bg-accent"
                                    >
                                        <Checkbox
                                            id={id}
                                            checked={selected.has(member.userId)}
                                            onCheckedChange={checked => {
                                                toggle(member.userId, checked === true);
                                            }}
                                        />
                                        <span className="flex min-w-0 flex-1 flex-col">
                                            <span className="truncate text-sm font-medium">
                                                {member.fullName || member.username}
                                            </span>
                                            {member.fullName ? (
                                                <span className="truncate text-xs text-muted-foreground">
                                                    {member.username}
                                                </span>
                                            ) : null}
                                        </span>
                                        {member.status === EOrgMemberStatus.Suspended && (
                                            <span className="shrink-0 text-xs text-muted-foreground">Suspended</span>
                                        )}
                                    </label>
                                </li>
                            );
                        })}
                    </ul>
                )}
            </DialogBody>
            {footer}
        </form>
    );
}

interface Props {
    /** Who may be in the group: the organization's members who are not removed. */
    members: readonly OrgMember[];
    /** Who is in it now. */
    defaultUserIds: readonly string[];
    /** Everybody to be in it. */
    onSubmit: (userIds: string[]) => void;
    onHasChanges?: (dirty: boolean) => void;
    footer?: React.ReactNode;
}
