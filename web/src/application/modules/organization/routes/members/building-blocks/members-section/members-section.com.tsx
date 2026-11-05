import { useCallback, useId, useMemo, useState } from "react";

import { Checkbox } from "@components/ui/checkbox";
import { Label } from "@components/ui/label";
import { OrgMembersQueries } from "~/organization/data/queries";
import { EOrgMemberStatus, type OrgMember } from "~/organization/domain";
import { getOrgMemberAbilities } from "~/organization/module-shared/utils";

import { ContentBlock, TableActions } from "@application/shared/components";
import { EOrgRole } from "@application/shared/enums";
import { useTableState } from "@application/shared/hooks/table";

import { DataTable } from "@/components/ui";

import { MembersTableDefs } from "./members-table.defs";

export function MembersSection({ orgId, myUserId, myRole }: Props) {
    const showRemovedId = useId();
    const { search, setSearch } = useTableState();
    const [showRemoved, setShowRemoved] = useState(false);
    const { data, isFetching } = OrgMembersQueries.useFindMany({ orgId });

    const all = useMemo(() => data?.data ?? [], [data]);
    const removedCount = all.filter(member => member.status === EOrgMemberStatus.Removed).length;

    const getAbilities = useCallback(
        (member: OrgMember) => getOrgMemberAbilities({ userId: myUserId, role: myRole }, member, all),
        [myUserId, myRole, all],
    );

    const columns = useMemo(
        () => MembersTableDefs.columns({ orgId, actorIsOwner: myRole === EOrgRole.Owner, getAbilities }),
        [orgId, myRole, getAbilities],
    );

    // The server sends every member at once, sorted by username: the search narrows them here. Removed members are
    // left out unless asked for.
    const members = useMemo(() => {
        const term = search.trim().toLowerCase();

        return all.filter(member => {
            if (!showRemoved && member.status === EOrgMemberStatus.Removed) {
                return false;
            }

            if (!term) {
                return true;
            }

            return (
                member.username.toLowerCase().includes(term) ||
                member.fullName.toLowerCase().includes(term) ||
                member.email.toLowerCase().includes(term)
            );
        });
    }, [all, search, showRemoved]);

    return (
        <ContentBlock label="Members">
            <div className="flex flex-col gap-4">
                <div className="flex flex-col gap-1 text-sm text-muted-foreground">
                    <p>
                        Owners and admins manage the organization. Only an owner can make or change an owner, and the
                        organization always keeps one active owner.
                    </p>
                    <p>
                        A suspended member gets nothing from the organization until reactivated, and keeps their role
                        and SSH accounts. A removed member loses both, and can be added back as a member.
                    </p>
                </div>
                <TableActions
                    search={{ value: search, onChange: setSearch }}
                    renderAfterSearch={
                        removedCount > 0 ? (
                            <div className="flex shrink-0 items-center gap-2">
                                <Checkbox
                                    id={showRemovedId}
                                    checked={showRemoved}
                                    onCheckedChange={checked => {
                                        setShowRemoved(checked === true);
                                    }}
                                />
                                <Label
                                    htmlFor={showRemovedId}
                                    className="font-normal"
                                >
                                    Show removed ({removedCount})
                                </Label>
                            </div>
                        ) : null
                    }
                />
                <DataTable
                    columns={columns}
                    data={members}
                    getRowId={member => member.userId}
                    rowClassName={member => (member.status === EOrgMemberStatus.Removed ? "opacity-60" : "")}
                    enablePagination
                    enableSorting
                    isLoading={isFetching}
                />
            </div>
        </ContentBlock>
    );
}

interface Props {
    orgId: string;
    /** The signed-in user, and their role in the organization: what they may do to whom. */
    myUserId: string;
    myRole: string;
}
