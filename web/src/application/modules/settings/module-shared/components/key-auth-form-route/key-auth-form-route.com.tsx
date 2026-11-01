import { useState } from "react";

import { toast } from "sonner";
import { KeyAuthCommands } from "~/settings/data/commands";
import { KeyAuthQueries } from "~/settings/data/queries";
import type { SettingKeyAuth } from "~/settings/domain";
import { CreateOrEditKeyAuthForm } from "~/settings/module-shared/components/key-auth-form";
import type {
    CreateOrEditKeyAuthFormInput,
    CreateOrEditKeyAuthFormOutput,
} from "~/settings/module-shared/components/key-auth-form";
import { SettingsFormRouteHeader } from "~/settings/module-shared/components/settings-form-route-header";
import { useSettingRevealSecrets } from "~/settings/module-shared/hooks";

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalModule } from "@application/shared/permissions";

import { RevealSecretsProvider } from "@/components/ui/input-password";

import { ConfirmRevealSecretsDialog } from "../confirm-reveal-secrets-dialog";
import { RevealSecretsButton } from "../reveal-secrets-button";

type KeyAuthFormRouteMode = "create" | "edit";

export function KeyAuthFormRoute({ mode, keyAuthId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const [saveRevision, setSaveRevision] = useState(0);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Settings });
    const { navigate } = useAppNavigate();

    const isEditMode = mode === "edit";
    const detailId = isEditMode ? (keyAuthId ?? "") : "";

    function navigateToList() {
        navigate.modules(ROUTE.settings.keyAuth.$route, { ignorePrevPath: true });
    }

    function markSaved() {
        setHasChanges(false);
        setSaveRevision(revision => revision + 1);
        navigateToList();
    }

    const { mutate: createKeyAuth, isPending: isCreating } = KeyAuthCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Key auth created successfully");
            markSaved();
        },
    });
    const { mutate: updateKeyAuth, isPending: isUpdating } = KeyAuthCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Key auth updated successfully");
            markSaved();
        },
    });

    const detailQuery = KeyAuthQueries.useFindOneById({ id: detailId }, { enabled: isEditMode });
    const keyAuth = detailQuery.data?.data;

    function createPayload(values: CreateOrEditKeyAuthFormOutput) {
        return {
            inheritable: values.inheritable,
            default: values.default,
            name: values.name,
            keyId: values.keyId,
            secretKey: values.secretKey,
        };
    }

    function onSubmit(values: CreateOrEditKeyAuthFormOutput) {
        const payload = createPayload(values);

        if (isEditMode && keyAuth) {
            updateKeyAuth({ id: keyAuth.id, payload: { ...payload, updateVer: keyAuth.updateVer } });
            return;
        }

        createKeyAuth({ payload });
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
    } = useSettingRevealSecrets<SettingKeyAuth>({
        settingType: "key-auth",
        settingId: detailId,
        mode,
    });

    const activeKeyAuth = revealedData ?? keyAuth;

    const isPending = isCreating || isUpdating;
    const isDetailLoading = isEditMode && detailQuery.isFetching;
    const initialValues: Partial<CreateOrEditKeyAuthFormInput> | undefined = activeKeyAuth
        ? {
              name: activeKeyAuth.name,
              keyId: activeKeyAuth.keyId,
              secretKey: activeKeyAuth.secretKey,
              inheritable: Boolean(activeKeyAuth.inheritable),
              default: activeKeyAuth.default ?? false,
          }
        : undefined;
    const shouldRenderForm = mode === "create" || Boolean(initialValues);
    const title = mode === "create" ? "Create Key Auth" : "Edit Key Auth";

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
                    <CreateOrEditKeyAuthForm
                        key={`${detailId}-${revealRevision}`}
                        isPending={isPending}
                        onSubmit={onSubmit}
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
    mode: KeyAuthFormRouteMode;
    keyAuthId?: string;
}
