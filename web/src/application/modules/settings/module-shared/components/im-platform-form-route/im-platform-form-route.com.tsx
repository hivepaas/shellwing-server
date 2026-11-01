import { useState } from "react";

import { toast } from "sonner";
import { ImServiceCommands } from "~/settings/data/commands";
import { ImServiceQueries } from "~/settings/data/queries";
import type { SettingImService } from "~/settings/domain";
import { CreateOrEditImPlatformForm } from "~/settings/module-shared/components/im-platform-form";
import type {
    CreateOrEditImPlatformFormInput,
    CreateOrEditImPlatformFormOutput,
} from "~/settings/module-shared/components/im-platform-form";
import { SettingsFormRouteHeader } from "~/settings/module-shared/components/settings-form-route-header";
import { useSettingRevealSecrets } from "~/settings/module-shared/hooks";

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { EImServiceKind } from "@application/shared/enums";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalModule } from "@application/shared/permissions";

import { RevealSecretsProvider } from "@/components/ui/input-password";

import { ConfirmRevealSecretsDialog } from "../confirm-reveal-secrets-dialog";
import { RevealSecretsButton } from "../reveal-secrets-button";

type ImPlatformFormRouteMode = "create" | "edit";

export function ImPlatformFormRoute({ mode, imPlatformId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const [saveRevision, setSaveRevision] = useState(0);
    const [testStatus, setTestStatus] = useState<"idle" | "succeeded" | "failed">("idle");
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Settings });
    const { navigate } = useAppNavigate();

    const isEditMode = mode === "edit";
    const detailId = isEditMode ? (imPlatformId ?? "") : "";

    function navigateToList() {
        navigate.modules(ROUTE.settings.imPlatforms.$route, { ignorePrevPath: true });
    }

    function markSaved() {
        setHasChanges(false);
        setSaveRevision(revision => revision + 1);
        navigateToList();
    }

    const { mutate: createImPlatform, isPending: isCreating } = ImServiceCommands.useCreateOne({
        onSuccess: () => {
            toast.success("IM platform created successfully");
            markSaved();
        },
    });
    const { mutate: updateImPlatform, isPending: isUpdating } = ImServiceCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("IM platform updated successfully");
            markSaved();
        },
    });
    const { mutate: testSendMsg, isPending: isTesting } = ImServiceCommands.useTestSendMsg({
        onSuccess: () => {
            setTestStatus("succeeded");
        },
        onError: () => {
            setTestStatus("failed");
        },
    });

    const detailQuery = ImServiceQueries.useFindOneById({ id: detailId }, { enabled: isEditMode });
    const imPlatform = detailQuery.data?.data;

    function createPayload(values: CreateOrEditImPlatformFormOutput) {
        const basePayload = {
            inheritable: values.inheritable,
            default: values.default,
            name: values.name,
            kind: values.kind,
        };

        if (values.kind === EImServiceKind.Slack) {
            return {
                ...basePayload,
                slack: { webhook: values.webhook },
                discord: null,
            };
        }

        return {
            ...basePayload,
            slack: null,
            discord: { webhook: values.webhook },
        };
    }

    function onSubmit(values: CreateOrEditImPlatformFormOutput) {
        const payload = createPayload(values);

        if (isEditMode && imPlatform) {
            updateImPlatform({ id: imPlatform.id, payload: { ...payload, updateVer: imPlatform.updateVer } });
            return;
        }

        createImPlatform({ payload });
    }

    function onTestSendMsg(values: CreateOrEditImPlatformFormOutput) {
        setTestStatus("idle");
        testSendMsg({
            payload: {
                ...createPayload(values),
                testMsg: "test message",
            },
        });
    }

    function handleClose() {
        if (isPending) {
            return;
        }
        if (canWrite && hasChanges && !window.confirm("Are you sure you want to close without saving changes?")) {
            return;
        }

        navigateToList();
    }

    const {
        canShowRevealButton,
        isDialogOpen,
        setIsDialogOpen,
        isRevealing,
        isRevealed,
        revealedData,
        revealRevision,
        handleConfirmReveal,
    } = useSettingRevealSecrets<SettingImService>({
        settingType: "im-services",
        settingId: detailId,
        mode,
    });

    const activeImPlatform = revealedData ?? imPlatform;

    const isPending = isCreating || isUpdating;
    const isDetailLoading = isEditMode && detailQuery.isFetching;
    const initialValues: Partial<CreateOrEditImPlatformFormInput> | undefined = activeImPlatform
        ? {
              name: activeImPlatform.name,
              kind: activeImPlatform.kind,
              webhook:
                  activeImPlatform.kind === EImServiceKind.Slack
                      ? (activeImPlatform.slack?.webhook ?? "")
                      : (activeImPlatform.discord?.webhook ?? ""),
              inheritable: Boolean(activeImPlatform.inheritable),
              default: activeImPlatform.default ?? false,
          }
        : undefined;
    const shouldRenderForm = mode === "create" || Boolean(initialValues);
    const title = mode === "create" ? "Create IM Platform" : "Edit IM Platform";

    return (
        <div className="flex w-full flex-col">
            <SettingsFormRouteHeader
                title={title}
                actions={
                    canShowRevealButton ? (
                        <RevealSecretsButton
                            onClick={() => {
                                setIsDialogOpen(true);
                            }}
                            isLoading={isRevealing}
                        />
                    ) : undefined
                }
            />

            <ConfirmRevealSecretsDialog
                open={isDialogOpen}
                onOpenChange={setIsDialogOpen}
                onConfirm={handleConfirmReveal}
                isPending={isRevealing}
            />

            {isDetailLoading && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}

            {!isDetailLoading && shouldRenderForm && (
                <RevealSecretsProvider value={{ isRevealed }}>
                    <CreateOrEditImPlatformForm
                        key={`${detailId}-${revealRevision}`}
                        isPending={isPending}
                        isTesting={isTesting}
                        testStatus={testStatus}
                        onSubmit={onSubmit}
                        onTestSendMsg={onTestSendMsg}
                        onHasChanges={setHasChanges}
                        savedVersion={saveRevision}
                        initialValues={initialValues}
                        showAvailableInProjects
                        readOnly={!canWrite}
                        onClose={handleClose}
                    />
                </RevealSecretsProvider>
            )}
        </div>
    );
}

interface Props {
    mode: ImPlatformFormRouteMode;
    imPlatformId?: string;
}
