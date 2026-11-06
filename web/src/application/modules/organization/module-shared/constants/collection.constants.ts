import { ECollectionPermission } from "~/organization/domain";

/**
 * What each permission on a collection lets one do, from the least to the most: each includes the ones before.
 */
export const COLLECTION_PERMISSION_OPTIONS: readonly { value: ECollectionPermission; label: string; hint: string }[] = [
    { value: ECollectionPermission.Connect, label: "Connect", hint: "See the hosts and connect" },
    { value: ECollectionPermission.Edit, label: "Edit", hint: "Also add and change them" },
    { value: ECollectionPermission.Manage, label: "Manage", hint: "Also grant others and replace the key" },
];

export function isCollectionPermission(value: string): value is ECollectionPermission {
    return COLLECTION_PERMISSION_OPTIONS.some(option => option.value === value);
}

/** The permission's name as shown, or the value itself when it is not a known one. */
export function getCollectionPermissionLabel(value: string): string {
    return COLLECTION_PERMISSION_OPTIONS.find(option => option.value === value)?.label ?? value;
}
