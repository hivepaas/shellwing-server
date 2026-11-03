import { useMemo } from "react";

import { OrgMembersQueries } from "~/organization/data/queries";

import { ContentBlock, TableActions } from "@application/shared/components";
import { useTableState } from "@application/shared/hooks/table";

import { DataTable } from "@/components/ui";

import { SshAccessTableDefs } from "./ssh-access-table.defs";

export function SshAccessSection({ orgId }: Props) {
    const { search, setSearch } = useTableState();
    const { data, isFetching } = OrgMembersQueries.useFindMany({ orgId });
    const columns = useMemo(() => SshAccessTableDefs.columns(orgId), [orgId]);

    // The server sends every member at once, sorted by username: the search narrows them here.
    const members = useMemo(() => {
        const all = data?.data ?? [];
        const term = search.trim().toLowerCase();

        if (!term) {
            return all;
        }

        return all.filter(
            member =>
                member.username.toLowerCase().includes(term) ||
                member.fullName.toLowerCase().includes(term) ||
                member.email.toLowerCase().includes(term) ||
                member.sshPrincipals.some(name => name.toLowerCase().includes(term)),
        );
    }, [data, search]);

    return (
        <ContentBlock label="Who may log in as what">
            <div className="flex flex-col gap-4">
                <div className="flex flex-col gap-1 text-sm text-muted-foreground">
                    <p>
                        The accounts each member&apos;s certificates may log in as on your servers, such as{" "}
                        <code className="font-mono text-xs">deploy</code>. Nobody has any until granted here, owners
                        included. A username is never one by itself: whoever could take the name{" "}
                        <code className="font-mono text-xs">root</code> would get root on every server.
                    </p>
                    <p>A change applies from the member&apos;s next certificate.</p>
                </div>
                <TableActions search={{ value: search, onChange: setSearch }} />
                <DataTable
                    columns={columns}
                    data={members}
                    getRowId={member => member.userId}
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
}
