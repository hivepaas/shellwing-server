import { useState } from "react";

import { Button } from "@components/ui/button";
import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogDescription,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";
import { Skeleton } from "@components/ui/skeleton";
import { toast } from "sonner";
import type { Collections_GrantPayload } from "~/organization/api/services/collections-services";
import { CollectionsCommands } from "~/organization/data/commands";
import { CollectionsQueries, OrgGroupsQueries, OrgMembersQueries } from "~/organization/data/queries";

import { CollectionKeyNote, CollectionPeopleList } from "../building-blocks";
import { EditCollectionAccessForm } from "../form";
import { useEditCollectionAccessDialogState } from "../hooks";

/**
 * Who may use a collection, and where each of them stands with its key. It stays open once saved: who has access, and
 * who waits for the key, follow from the grants.
 */
export function EditCollectionAccessDialog() {
    const { state, props: dialogProps, ...actions } = useEditCollectionAccessDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    /** How many times the grants were saved here: the form starts again from the saved grants each time. */
    const [savedCount, setSavedCount] = useState(0);

    const open = state.mode !== "closed";
    const orgId = state.orgId ?? "";
    const collectionId = state.collection?.id ?? "";

    // Not kept once closed: the form starts from the grants read as it opens, not from an older read.
    const detailQuery = CollectionsQueries.useFindOne({ orgId, collectionId }, { enabled: open, gcTime: 0 });
    const membersQuery = OrgMembersQueries.useFindMany({ orgId }, { enabled: open });
    const groupsQuery = OrgGroupsQueries.useFindMany({ orgId }, { enabled: open });

    const detail = detailQuery.data?.data;
    const members = membersQuery.data?.data;
    const groups = groupsQuery.data?.data;

    const isLoaded = detail !== undefined && members !== undefined && groups !== undefined;
    const hasFailed =
        !isLoaded &&
        ((detail === undefined && detailQuery.isError) ||
            (members === undefined && membersQuery.isError) ||
            (groups === undefined && groupsQuery.isError));

    const {
        mutate: setGrants,
        isPending,
        reset,
    } = CollectionsCommands.useSetGrants({
        onSuccess: () => {
            toast.success("Access saved. Anybody newly granted waits for the key until an app holding it gives it.");
            setHasChanges(false);
            setSavedCount(count => count + 1);
            dialogProps?.onSuccess?.();
        },
    });

    function handleClose() {
        if (isPending) {
            return;
        }

        if (hasChanges) {
            const confirmed = window.confirm("Are you sure you want to close without saving changes?");

            if (!confirmed) {
                return;
            }
        }

        setHasChanges(false);
        reset();
        actions.close();
        dialogProps?.onClose?.();
    }

    function onSubmit(grants: Collections_GrantPayload[]) {
        if (state.mode !== "open") {
            return;
        }

        setGrants({
            orgId: state.orgId,
            collectionId: state.collection.id,
            payload: { grants },
        });
    }

    const collection = detail?.collection ?? state.collection;

    return (
        <Dialog
            open={open}
            modal
            onOpenChange={handleClose}
        >
            <DialogFixedContent className="sm:max-w-[760px]">
                <DialogHeader>
                    <DialogTitle>Access to {collection?.name ?? ""}</DialogTitle>
                    <DialogDescription>
                        Grant people and groups Connect, Edit or Manage. Saving replaces every grant.
                    </DialogDescription>
                </DialogHeader>

                {hasFailed && (
                    <DialogBody className="text-sm text-muted-foreground">
                        The collection, or the members and groups of the organization, could not be read. Close this and
                        try again.
                    </DialogBody>
                )}

                {!isLoaded && !hasFailed && (
                    <DialogBody className="flex flex-col gap-3">
                        <Skeleton className="h-16 w-full" />
                        <Skeleton className="h-9 w-full" />
                        <Skeleton className="h-40 w-full" />
                    </DialogBody>
                )}

                {isLoaded && (
                    <EditCollectionAccessForm
                        key={`${detail.collection.id}:${savedCount}`}
                        defaultGrants={detail.grants}
                        members={members}
                        groups={groups}
                        before={<CollectionKeyNote collection={detail.collection} />}
                        after={
                            <section className="flex flex-col gap-3">
                                <div className="flex flex-col gap-1">
                                    <h3 className="text-sm font-semibold">People with access</h3>
                                    <span className="text-xs text-muted-foreground">
                                        As saved: through their role, their grant or their groups&apos;. Compare the
                                        fingerprint the shellwing app shows with the person before giving them the key.
                                    </span>
                                </div>
                                <CollectionPeopleList members={detail.members} />
                            </section>
                        }
                        onSubmit={onSubmit}
                        onHasChanges={setHasChanges}
                        footer={
                            <DialogActionFooter className="flex items-center justify-end gap-3">
                                <Button
                                    type="button"
                                    variant="outline"
                                    className="min-w-[100px]"
                                    disabled={isPending}
                                    onClick={handleClose}
                                >
                                    Close
                                </Button>
                                <Button
                                    type="submit"
                                    variant="default"
                                    className="min-w-[100px]"
                                    disabled={!hasChanges}
                                    isLoading={isPending}
                                >
                                    Save
                                </Button>
                            </DialogActionFooter>
                        }
                    />
                )}
            </DialogFixedContent>
        </Dialog>
    );
}
