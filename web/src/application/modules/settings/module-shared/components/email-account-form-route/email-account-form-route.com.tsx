import { useState } from "react";

import { toast } from "sonner";
import { EmailCommands } from "~/settings/data/commands";
import { EmailQueries } from "~/settings/data/queries";
import type { SettingEmail } from "~/settings/domain";
import {
    CreateOrEditEmailAccountForm,
    TestSendMailDialog,
} from "~/settings/module-shared/components/email-account-form";
import type {
    CreateOrEditEmailAccountFormInput,
    CreateOrEditEmailAccountFormOutput,
    TestSendMailFormOutput,
} from "~/settings/module-shared/components/email-account-form";
import { SettingsFormRouteHeader } from "~/settings/module-shared/components/settings-form-route-header";
import { useSettingRevealSecrets } from "~/settings/module-shared/hooks";

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { EEmailKind } from "@application/shared/enums";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalModule } from "@application/shared/permissions";

import { RevealSecretsProvider } from "@/components/ui/input-password";

import { ConfirmRevealSecretsDialog } from "../confirm-reveal-secrets-dialog";
import { RevealSecretsButton } from "../reveal-secrets-button";

type EmailAccountFormRouteMode = "create" | "edit";

const FIELD_MAPPING_NAMES = [
    "fromAddress",
    "fromName",
    "toAddress",
    "toAddresses",
    "subject",
    "content",
    "password",
] as const;

export function EmailAccountFormRoute({ mode, emailAccountId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const [saveRevision, setSaveRevision] = useState(0);
    const [testStatus, setTestStatus] = useState<"idle" | "succeeded" | "failed">("idle");
    const [testDialogOpen, setTestDialogOpen] = useState(false);
    const [testAccountValues, setTestAccountValues] = useState<CreateOrEditEmailAccountFormOutput | null>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Settings });
    const { navigate } = useAppNavigate();

    const isEditMode = mode === "edit";
    const detailId = isEditMode ? (emailAccountId ?? "") : "";

    function navigateToList() {
        navigate.modules(ROUTE.settings.emailAccounts.$route, { ignorePrevPath: true });
    }

    function markSaved() {
        setHasChanges(false);
        setSaveRevision(revision => revision + 1);
        navigateToList();
    }

    const { mutate: createEmailAccount, isPending: isCreating } = EmailCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Email account created successfully");
            markSaved();
        },
    });
    const { mutate: updateEmailAccount, isPending: isUpdating } = EmailCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Email account updated successfully");
            markSaved();
        },
    });
    const { mutate: testSendMail, isPending: isTesting } = EmailCommands.useTestSendMail({
        onSuccess: () => {
            setTestStatus("succeeded");
        },
        onError: () => {
            setTestStatus("failed");
        },
    });

    const detailQuery = EmailQueries.useFindOneById({ id: detailId }, { enabled: isEditMode });
    const emailAccount = detailQuery.data?.data;

    function createPayload(values: CreateOrEditEmailAccountFormOutput) {
        const isSmtp = values.kind === EEmailKind.SMTP;
        const headers = toRecord(values.headers);

        return {
            inheritable: values.inheritable,
            default: values.default,
            name: values.name,
            kind: values.kind,
            smtp: isSmtp
                ? {
                      host: values.smtpHost,
                      port: values.smtpPort,
                      username: values.smtpUsername,
                      displayName: values.smtpDisplayName,
                      password: values.smtpPassword,
                      ssl: false,
                  }
                : null,
            http: isSmtp
                ? null
                : {
                      endpoint: values.httpEndpoint,
                      method: values.httpMethod,
                      contentType: values.httpContentType,
                      headers,
                      fieldMapping: toFieldMapping(values.fieldMapping),
                      username: values.httpUsername,
                      displayName: values.httpDisplayName,
                      password: values.httpPassword,
                  },
        };
    }

    function onSubmit(values: CreateOrEditEmailAccountFormOutput) {
        const payload = createPayload(values);

        if (isEditMode && emailAccount) {
            updateEmailAccount({ id: emailAccount.id, payload: { ...payload, updateVer: emailAccount.updateVer } });
            return;
        }

        createEmailAccount({ payload });
    }

    function openTestSendMailDialog(values: CreateOrEditEmailAccountFormOutput) {
        setTestStatus("idle");
        setTestAccountValues(values);
        setTestDialogOpen(true);
    }

    function onTestSendMail(values: TestSendMailFormOutput) {
        if (!testAccountValues) {
            return;
        }

        setTestStatus("idle");
        const { inheritable: _inheritable, default: _default, ...payload } = createPayload(testAccountValues);

        testSendMail({
            payload: {
                ...payload,
                testRecipient: values.testRecipient,
                testSubject: values.testSubject,
                testContent: values.testContent,
            },
        });
    }

    function handleClose() {
        if (isPending || isTesting) {
            return;
        }
        if (canWrite && hasChanges && !window.confirm("Are you sure you want to close without saving changes?")) {
            return;
        }

        navigateToList();
    }

    function handleTestDialogOpenChange(nextOpen: boolean) {
        if (isTesting) {
            return;
        }

        setTestDialogOpen(nextOpen);

        if (!nextOpen) {
            setTestStatus("idle");
            setTestAccountValues(null);
        }
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
    } = useSettingRevealSecrets<SettingEmail>({
        settingType: "emails",
        settingId: detailId,
        mode,
    });

    const activeEmailAccount = revealedData ?? emailAccount;

    const isPending = isCreating || isUpdating;
    const isDetailLoading = isEditMode && detailQuery.isFetching;
    const initialValues: Partial<CreateOrEditEmailAccountFormInput> | undefined = activeEmailAccount
        ? toInitialValues(activeEmailAccount)
        : undefined;
    const shouldRenderForm = mode === "create" || Boolean(initialValues);
    const title = mode === "create" ? "Create Email Account" : "Edit Email Account";

    return (
        <>
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
                        <CreateOrEditEmailAccountForm
                            key={`${detailId}-${revealRevision}`}
                            isPending={isPending}
                            onSubmit={onSubmit}
                            onOpenTestSendMail={openTestSendMailDialog}
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
            <TestSendMailDialog
                open={testDialogOpen}
                isPending={isTesting}
                testStatus={testStatus}
                onOpenChange={handleTestDialogOpenChange}
                onSubmit={onTestSendMail}
            />
        </>
    );
}

function toKeyValueList(map?: object | null) {
    return Object.entries(map ?? {}).map(([key, value]) => ({ key, value: String(value ?? "") }));
}

function toRecord(rows: { key: string; value: string }[]) {
    return rows.reduce<Record<string, string>>((acc, item) => {
        const key = item.key.trim();
        if (!key) {
            return acc;
        }

        return {
            ...acc,
            [key]: item.value.trim(),
        };
    }, {});
}

function toFieldMapping(rows: { key: string; value: string }[]) {
    const mapping: Record<(typeof FIELD_MAPPING_NAMES)[number], string> = {
        fromAddress: "",
        fromName: "",
        toAddress: "",
        toAddresses: "",
        subject: "",
        content: "",
        password: "",
    };

    return rows.reduce((acc, item) => {
        const key = item.key.trim();
        if (!FIELD_MAPPING_NAMES.includes(key as (typeof FIELD_MAPPING_NAMES)[number])) {
            return acc;
        }

        return {
            ...acc,
            [key]: item.value.trim(),
        };
    }, mapping);
}

function toInitialValues(emailAccount: SettingEmail): Partial<CreateOrEditEmailAccountFormOutput> {
    return {
        name: emailAccount.name,
        kind: emailAccount.kind,
        smtpHost: emailAccount.smtp?.host ?? "",
        smtpPort: emailAccount.smtp?.port ?? 587,
        smtpUsername: emailAccount.smtp?.username ?? "",
        smtpDisplayName: emailAccount.smtp?.displayName ?? "",
        smtpPassword: emailAccount.smtp?.password ?? "",
        httpEndpoint: emailAccount.http?.endpoint ?? "",
        httpMethod:
            emailAccount.http?.method === "GET" || emailAccount.http?.method === "PUT"
                ? emailAccount.http.method
                : "POST",
        httpUsername: emailAccount.http?.username ?? "",
        httpDisplayName: emailAccount.http?.displayName ?? "",
        httpPassword: emailAccount.http?.password ?? "",
        httpContentType: emailAccount.http?.contentType ?? "application/json",
        headers: toKeyValueList(emailAccount.http?.headers),
        fieldMapping: toKeyValueList(emailAccount.http?.fieldMapping),
        inheritable: Boolean(emailAccount.inheritable),
        default: emailAccount.default ?? false,
    };
}

interface Props {
    mode: EmailAccountFormRouteMode;
    emailAccountId?: string;
}
