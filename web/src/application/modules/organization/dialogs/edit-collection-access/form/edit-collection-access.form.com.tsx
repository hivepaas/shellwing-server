import React, { useMemo, useState } from "react";

import { Button } from "@components/ui/button";
import { DialogBody } from "@components/ui/dialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { Group, User, X } from "lucide-react";
import type { Collections_GrantPayload } from "~/organization/api/services/collections-services";
import {
    type CollectionGrant,
    ECollectionPermission,
    EOrgMemberStatus,
    type OrgGroup,
    type OrgMember,
} from "~/organization/domain";
import {
    COLLECTION_PERMISSION_OPTIONS,
    getCollectionPermissionLabel,
    isCollectionPermission,
} from "~/organization/module-shared/constants";

import { EOrgRole } from "@application/shared/enums";

/**
 * A grant as it is edited: to a person or to a group, by their id.
 */
interface GrantRow {
    /** "user:{id}" or "group:{id}": a person or a group is granted once. */
    key: string;
    kind: "user" | "group";
    id: string;
    name: string;
    /** Below the name: a person's username, or that it is a group. */
    detail: string;
    permission: ECollectionPermission;
}

function toRow(grant: CollectionGrant): GrantRow {
    const permission = isCollectionPermission(grant.permission) ? grant.permission : ECollectionPermission.Connect;

    if (grant.groupId) {
        return {
            key: `group:${grant.groupId}`,
            kind: "group",
            id: grant.groupId,
            name: grant.groupName || grant.groupId,
            detail: "Group",
            permission,
        };
    }

    return {
        key: `user:${grant.userId}`,
        kind: "user",
        id: grant.userId,
        name: grant.fullName || grant.username || grant.userId,
        detail: grant.fullName ? grant.username : "",
        permission,
    };
}

function toPayload(row: GrantRow): Collections_GrantPayload {
    return row.kind === "user"
        ? { userId: row.id, permission: row.permission }
        : { groupId: row.id, permission: row.permission };
}

/** Below a member's name where they are picked: their username, and what changes what a grant gives them. */
function getMemberDetail(member: OrgMember): string {
    return [
        member.fullName ? member.username : "",
        member.role === EOrgRole.Owner ? "Owner: has Manage already" : "",
        member.status === EOrgMemberStatus.Suspended ? "Suspended" : "",
    ]
        .filter(Boolean)
        .join(" · ");
}

function isSameGrants(rows: readonly GrantRow[], defaults: readonly GrantRow[]): boolean {
    if (rows.length !== defaults.length) {
        return false;
    }

    const byKey = new Map(defaults.map(row => [row.key, row.permission]));

    return rows.every(row => byKey.get(row.key) === row.permission);
}

/**
 * The collection's grants: each a person or a group with a permission. People and groups are added from the
 * organization's members who are not removed, and its groups. Saving sends them all, in place of what was.
 */
export function EditCollectionAccessForm({
    defaultGrants,
    members,
    groups,
    before,
    after,
    onSubmit,
    onHasChanges,
    footer,
}: Props) {
    const defaultRows = useMemo(() => defaultGrants.map(toRow), [defaultGrants]);
    const [rows, setRows] = useState<GrantRow[]>(defaultRows);

    const isDirty = useMemo(() => !isSameGrants(rows, defaultRows), [rows, defaultRows]);

    React.useEffect(() => {
        onHasChanges?.(isDirty);
    }, [isDirty, onHasChanges]);

    const granted = useMemo(() => new Set(rows.map(row => row.key)), [rows]);

    const availableMembers = useMemo(
        () =>
            members.filter(
                member => member.status !== EOrgMemberStatus.Removed && !granted.has(`user:${member.userId}`),
            ),
        [members, granted],
    );

    const availableGroups = useMemo(() => groups.filter(group => !granted.has(`group:${group.id}`)), [groups, granted]);

    function addMember(userId: string) {
        const member = members.find(m => m.userId === userId);

        if (member === undefined) {
            return;
        }

        setRows(previous => [
            ...previous,
            {
                key: `user:${member.userId}`,
                kind: "user",
                id: member.userId,
                name: member.fullName || member.username,
                detail: member.fullName ? member.username : "",
                permission: ECollectionPermission.Connect,
            },
        ]);
    }

    function addGroup(groupId: string) {
        const group = groups.find(g => g.id === groupId);

        if (group === undefined) {
            return;
        }

        setRows(previous => [
            ...previous,
            {
                key: `group:${group.id}`,
                kind: "group",
                id: group.id,
                name: group.name,
                detail: "Group",
                permission: ECollectionPermission.Connect,
            },
        ]);
    }

    function setPermission(key: string, permission: ECollectionPermission) {
        setRows(previous => previous.map(row => (row.key === key ? { ...row, permission } : row)));
    }

    function remove(key: string) {
        setRows(previous => previous.filter(row => row.key !== key));
    }

    return (
        <form
            onSubmit={event => {
                event.preventDefault();
                onSubmit(rows.map(toPayload));
            }}
            className="min-h-0 flex flex-1 flex-col"
        >
            <DialogBody className="flex flex-col gap-6">
                {before}

                <section className="flex flex-col gap-3">
                    <div className="flex flex-col gap-1">
                        <h3 className="text-sm font-semibold">Granted</h3>
                        <ul className="flex flex-col gap-0.5 text-xs text-muted-foreground">
                            {COLLECTION_PERMISSION_OPTIONS.map(option => (
                                <li key={option.value}>
                                    <span className="font-medium text-foreground">{option.label}</span>: {option.hint}
                                </li>
                            ))}
                        </ul>
                        <span className="text-xs text-muted-foreground">
                            Owners have Manage on every collection without being granted.
                        </span>
                    </div>

                    {rows.length === 0 ? (
                        <span className="rounded-md border border-dashed px-3 py-4 text-center text-sm text-muted-foreground">
                            Nobody is granted yet. Add a person or a group.
                        </span>
                    ) : (
                        <ul className="flex flex-col rounded-md border">
                            {rows.map(row => {
                                const Icon = row.kind === "group" ? Group : User;

                                return (
                                    <li
                                        key={row.key}
                                        className="flex items-center gap-3 border-b px-3 py-2 last:border-b-0"
                                    >
                                        <Icon className="size-4 shrink-0 text-muted-foreground" />
                                        <span className="flex min-w-0 flex-1 flex-col">
                                            <span className="truncate text-sm font-medium">{row.name}</span>
                                            {row.detail ? (
                                                <span className="truncate text-xs text-muted-foreground">
                                                    {row.detail}
                                                </span>
                                            ) : null}
                                        </span>
                                        <Select
                                            value={row.permission}
                                            onValueChange={value => {
                                                if (isCollectionPermission(value)) {
                                                    setPermission(row.key, value);
                                                }
                                            }}
                                        >
                                            <SelectTrigger
                                                size="sm"
                                                className="w-[120px] shrink-0"
                                                aria-label={`Permission of ${row.name}`}
                                            >
                                                <SelectValue>
                                                    {getCollectionPermissionLabel(row.permission)}
                                                </SelectValue>
                                            </SelectTrigger>
                                            <SelectContent align="end">
                                                {COLLECTION_PERMISSION_OPTIONS.map(option => (
                                                    <SelectItem
                                                        key={option.value}
                                                        value={option.value}
                                                    >
                                                        <span className="flex flex-col">
                                                            <span>{option.label}</span>
                                                            <span className="text-xs text-muted-foreground">
                                                                {option.hint}
                                                            </span>
                                                        </span>
                                                    </SelectItem>
                                                ))}
                                            </SelectContent>
                                        </Select>
                                        <Button
                                            type="button"
                                            variant="ghost"
                                            size="icon"
                                            className="size-8 shrink-0"
                                            aria-label={`Remove ${row.name}`}
                                            onClick={() => {
                                                remove(row.key);
                                            }}
                                        >
                                            <X className="size-4" />
                                        </Button>
                                    </li>
                                );
                            })}
                        </ul>
                    )}

                    <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
                        <Select
                            value=""
                            disabled={availableMembers.length === 0}
                            onValueChange={addMember}
                        >
                            <SelectTrigger aria-label="Add a person">
                                <SelectValue
                                    placeholder={
                                        availableMembers.length === 0 ? "No more people to add" : "Add a person"
                                    }
                                />
                            </SelectTrigger>
                            <SelectContent>
                                {availableMembers.map(member => {
                                    const detail = getMemberDetail(member);

                                    return (
                                        <SelectItem
                                            key={member.userId}
                                            value={member.userId}
                                        >
                                            <span className="flex flex-col">
                                                <span>{member.fullName || member.username}</span>
                                                {detail ? (
                                                    <span className="text-xs text-muted-foreground">{detail}</span>
                                                ) : null}
                                            </span>
                                        </SelectItem>
                                    );
                                })}
                            </SelectContent>
                        </Select>

                        <Select
                            value=""
                            disabled={availableGroups.length === 0}
                            onValueChange={addGroup}
                        >
                            <SelectTrigger aria-label="Add a group">
                                <SelectValue
                                    placeholder={
                                        groups.length === 0
                                            ? "The organization has no groups"
                                            : availableGroups.length === 0
                                              ? "No more groups to add"
                                              : "Add a group"
                                    }
                                />
                            </SelectTrigger>
                            <SelectContent>
                                {availableGroups.map(group => (
                                    <SelectItem
                                        key={group.id}
                                        value={group.id}
                                    >
                                        <span className="flex flex-col">
                                            <span>{group.name}</span>
                                            <span className="text-xs text-muted-foreground">
                                                {group.members.length === 1
                                                    ? "1 member"
                                                    : `${group.members.length} members`}
                                            </span>
                                        </span>
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </div>
                </section>

                {after}
            </DialogBody>
            {footer}
        </form>
    );
}

interface Props {
    /** The grants saved now. */
    defaultGrants: readonly CollectionGrant[];
    /** The organization's members: those not removed may be added. */
    members: readonly OrgMember[];
    /** The organization's groups. */
    groups: readonly OrgGroup[];
    /** Shown above the grants. */
    before?: React.ReactNode;
    /** Shown below the grants. */
    after?: React.ReactNode;
    /** Every grant of the collection. */
    onSubmit: (grants: Collections_GrantPayload[]) => void;
    onHasChanges?: (dirty: boolean) => void;
    footer?: React.ReactNode;
}
