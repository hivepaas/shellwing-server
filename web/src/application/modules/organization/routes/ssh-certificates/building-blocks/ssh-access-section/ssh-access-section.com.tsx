import { useCallback, useMemo } from "react";

import { OrgGroupsQueries, OrgMembersQueries } from "~/organization/data/queries";
import { EOrgMemberStatus, type OrgMember } from "~/organization/domain";
import { type GroupSshPrincipal, getGroupSshPrincipals } from "~/organization/module-shared/utils";

import { ContentBlock, TableActions } from "@application/shared/components";
import { useTableState } from "@application/shared/hooks/table";

import { DataTable } from "@/components/ui";

import { SshAccessTableDefs } from "./ssh-access-table.defs";

const NO_GROUP_PRINCIPALS: GroupSshPrincipal[] = [];

export function SshAccessSection({ orgId }: Props) {
    const { search, setSearch } = useTableState();
    const { data, isFetching } = OrgMembersQueries.useFindMany({ orgId });
    const { data: groupsData, isFetching: isFetchingGroups } = OrgGroupsQueries.useFindMany({ orgId });

    // Each member's accounts from their groups, besides their own.
    const groupPrincipals = useMemo(() => {
        const groupsById = new Map((groupsData?.data ?? []).map(group => [group.id, group]));

        return new Map(
            (data?.data ?? []).map(member => [member.userId, getGroupSshPrincipals(member, groupsById)] as const),
        );
    }, [data, groupsData]);

    const getGroupPrincipals = useCallback(
        (member: OrgMember) => groupPrincipals.get(member.userId) ?? NO_GROUP_PRINCIPALS,
        [groupPrincipals],
    );

    const columns = useMemo(
        () => SshAccessTableDefs.columns({ orgId, getGroupPrincipals }),
        [orgId, getGroupPrincipals],
    );

    // The server sends every member at once, sorted by username: the search narrows them here. A removed member has
    // no accounts, and can be given none until added back.
    const members = useMemo(() => {
        const all = (data?.data ?? []).filter(member => member.status !== EOrgMemberStatus.Removed);
        const term = search.trim().toLowerCase();

        if (!term) {
            return all;
        }

        return all.filter(
            member =>
                member.username.toLowerCase().includes(term) ||
                member.fullName.toLowerCase().includes(term) ||
                member.email.toLowerCase().includes(term) ||
                member.sshPrincipals.some(name => name.toLowerCase().includes(term)) ||
                getGroupPrincipals(member).some(({ name }) => name.toLowerCase().includes(term)),
        );
    }, [data, search, getGroupPrincipals]);

    return (
        <ContentBlock label="Who may log in as what">
            <div className="flex flex-col gap-4">
                <div className="flex flex-col gap-1 text-sm text-muted-foreground">
                    <p>
                        The accounts each member&apos;s certificates may log in as on your servers, such as{" "}
                        <code className="font-mono text-xs">deploy</code>. Nobody has any until granted here or through
                        a group, owners included. A username is never one by itself: whoever could take the name{" "}
                        <code className="font-mono text-xs">root</code> would get root on every server.
                    </p>
                    <p>
                        Dashed accounts come from the member&apos;s groups, and are changed under Groups. A change
                        applies from the member&apos;s next certificate.
                    </p>
                </div>
                <TableActions search={{ value: search, onChange: setSearch }} />
                <DataTable
                    columns={columns}
                    data={members}
                    getRowId={member => member.userId}
                    enablePagination
                    enableSorting
                    isLoading={isFetching || isFetchingGroups}
                />
            </div>
        </ContentBlock>
    );
}

interface Props {
    orgId: string;
}
