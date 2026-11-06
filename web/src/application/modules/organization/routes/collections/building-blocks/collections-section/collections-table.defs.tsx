import type { ColumnDef } from "@tanstack/react-table";
import type { Collection } from "~/organization/domain";
import { CollectionKeyBadges, CollectionPermissionBadge } from "~/organization/module-shared/components";

import { CollectionMenuCell, CollectionNameCell } from "./building-blocks";

function createColumns(orgId: string): ColumnDef<Collection>[] {
    return [
        {
            accessorKey: "name",
            header: "Collection",
            enableSorting: true,
            cell: ({ row: { original } }) => <CollectionNameCell collection={original} />,
        },
        {
            id: "key",
            accessorFn: collection => collection.keyVersion,
            header: "Key",
            enableSorting: true,
            cell: ({ row: { original } }) => <CollectionKeyBadges collection={original} />,
        },
        {
            id: "permission",
            accessorFn: collection => collection.permission,
            header: "Your access",
            enableSorting: false,
            cell: ({ row: { original } }) => <CollectionPermissionBadge permission={original.permission} />,
        },
        {
            id: "actions",
            header: "",
            enableSorting: false,
            minSize: 56,
            size: 56,
            cell: ({ row: { original } }) => (
                <CollectionMenuCell
                    orgId={orgId}
                    collection={original}
                />
            ),
            meta: {
                align: "right",
            },
        },
    ];
}

export const CollectionsTableDefs = Object.freeze({
    columns: createColumns,
});
