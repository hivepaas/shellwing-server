import { useState } from "react";

import { toast } from "sonner";
import { NotificationCommands } from "~/settings/data/commands";
import { NotificationQueries } from "~/settings/data/queries";
import { CreateOrEditNotificationTargetForm } from "~/settings/module-shared/components/notification-target-form";
import type {
    CreateOrEditNotificationTargetFormInput,
    CreateOrEditNotificationTargetFormOutput,
} from "~/settings/module-shared/components/notification-target-form";
import { SettingsFormRouteHeader } from "~/settings/module-shared/components/settings-form-route-header";

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalModule } from "@application/shared/permissions";

type NotificationTargetFormRouteMode = "create" | "edit";

function splitCommaSeparated(value: string): string[] {
    return value
        .split(",")
        .map(item => item.trim())
        .filter(Boolean);
}

function joinAddresses(value?: string[]): string {
    return value?.join(", ") ?? "";
}

export function NotificationTargetFormRoute({ mode, notificationTargetId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const [saveRevision, setSaveRevision] = useState(0);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Settings });
    const { navigate } = useAppNavigate();

    const isEditMode = mode === "edit";
    const detailId = isEditMode ? (notificationTargetId ?? "") : "";

    function navigateToList() {
        navigate.modules(ROUTE.settings.notificationTargets.$route, { ignorePrevPath: true });
    }

    function markSaved() {
        setHasChanges(false);
        setSaveRevision(revision => revision + 1);
        navigateToList();
    }

    const { mutate: createNotification, isPending: isCreating } = NotificationCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Notification target created successfully");
            markSaved();
        },
    });
    const { mutate: updateNotification, isPending: isUpdating } = NotificationCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Notification target updated successfully");
            markSaved();
        },
    });

    const detailQuery = NotificationQueries.useFindOneById({ id: detailId }, { enabled: isEditMode });
    const notificationTarget = detailQuery.data?.data;

    function createPayload(values: CreateOrEditNotificationTargetFormOutput) {
        const viaEmail = values.emailEnabled
            ? {
                  enabled: values.emailEnabled,
                  useDefault: values.emailUseDefault,
                  sender: { id: values.emailUseDefault ? "" : values.senderEmailAccountId },
                  toProjectMembers: values.notifyProjectMembers,
                  toProjectOwners: values.notifyProjectOwners,
                  toAllAdmins: values.notifyAdmins,
                  toAddresses: splitCommaSeparated(values.customAddresses),
              }
            : null;
        const viaSlack = values.slackEnabled
            ? {
                  enabled: values.slackEnabled,
                  useDefault: values.slackUseDefault,
                  webhook: { id: values.slackUseDefault ? "" : values.slackWebhookId },
              }
            : null;
        const viaDiscord = values.discordEnabled
            ? {
                  enabled: values.discordEnabled,
                  useDefault: values.discordUseDefault,
                  webhook: { id: values.discordUseDefault ? "" : values.discordWebhookId },
              }
            : null;

        return {
            inheritable: values.inheritable,
            default: values.default,
            name: values.name,
            viaEmail,
            viaSlack,
            viaDiscord,
            minSendInterval: values.minSendInterval,
        };
    }

    function onSubmit(values: CreateOrEditNotificationTargetFormOutput) {
        const payload = createPayload(values);

        if (isEditMode && notificationTarget) {
            updateNotification({
                id: notificationTarget.id,
                payload: { ...payload, updateVer: notificationTarget.updateVer },
            });
            return;
        }

        createNotification({ payload });
    }

    function handleClose() {
        if (isPending) return;
        if (canWrite && hasChanges && !window.confirm("Are you sure you want to close without saving changes?")) return;

        navigateToList();
    }

    const isPending = isCreating || isUpdating;
    const isDetailLoading = isEditMode && detailQuery.isFetching;
    const initialValues: Partial<CreateOrEditNotificationTargetFormInput> | undefined = notificationTarget
        ? {
              name: notificationTarget.name,
              emailEnabled: notificationTarget.viaEmail?.enabled ?? false,
              emailUseDefault: notificationTarget.viaEmail?.useDefault ?? true,
              senderEmailAccountId: notificationTarget.viaEmail?.sender?.id ?? "",
              notifyAdmins: notificationTarget.viaEmail?.toAllAdmins ?? false,
              notifyProjectOwners: notificationTarget.viaEmail?.toProjectOwners ?? false,
              notifyProjectMembers: notificationTarget.viaEmail?.toProjectMembers ?? false,
              customAddresses: joinAddresses(notificationTarget.viaEmail?.toAddresses),
              slackEnabled: notificationTarget.viaSlack?.enabled ?? false,
              slackUseDefault: notificationTarget.viaSlack?.useDefault ?? true,
              slackWebhookId: notificationTarget.viaSlack?.webhook?.id ?? "",
              discordEnabled: notificationTarget.viaDiscord?.enabled ?? false,
              discordUseDefault: notificationTarget.viaDiscord?.useDefault ?? true,
              discordWebhookId: notificationTarget.viaDiscord?.webhook?.id ?? "",
              minSendInterval: notificationTarget.minSendInterval || "3m",
              inheritable: Boolean(notificationTarget.inheritable),
              default: notificationTarget.default ?? false,
          }
        : undefined;
    const shouldRenderForm = mode === "create" || initialValues;
    const title = mode === "create" ? "Create Notification Target" : "Edit Notification Target";

    return (
        <div className="flex w-full flex-col">
            <SettingsFormRouteHeader title={title} />

            {isDetailLoading && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}

            {!isDetailLoading && shouldRenderForm && (
                <CreateOrEditNotificationTargetForm
                    isPending={isPending}
                    onSubmit={onSubmit}
                    onHasChanges={setHasChanges}
                    savedVersion={saveRevision}
                    initialValues={initialValues}
                    showAvailableInProjects
                    readOnly={!canWrite}
                    onClose={handleClose}
                />
            )}
        </div>
    );
}

interface Props {
    mode: NotificationTargetFormRouteMode;
    notificationTargetId?: string;
}
