import { useEffect, useMemo, useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import { EmailQueries, ImServiceQueries } from "~/settings/data/queries";
import type { SettingEmail, SettingImService } from "~/settings/domain";
import { SETTINGS_FORM_FIELD_CONTROL_MAX_WIDTH_CLASS } from "~/settings/module-shared/constants/settings-form-layout.constants";

import { AppLink, Combobox, FormActionBar, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { EImServiceKind } from "@application/shared/enums";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input } from "@/components/ui";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";

import { PermissionReadonlyNotice } from "../permission-readonly-notice.com";
import { SettingsFormCancelAction } from "../settings-form-cancel-action";

import type {
    CreateOrEditNotificationTargetFormInput,
    CreateOrEditNotificationTargetFormOutput,
} from "./create-or-edit-notification-target.form.schema";
import { CreateOrEditNotificationTargetFormSchema } from "./create-or-edit-notification-target.form.schema";

const selectPagination = { page: 1, size: 100 };
type NotificationSection = "email" | "slack" | "discord";

const initialOpenSections: NotificationSection[] = ["email", "slack", "discord"];

function mergeOpenSections(current: NotificationSection[], additions: NotificationSection[]) {
    return Array.from(new Set([...current, ...additions]));
}

function getInvalidSections(errors: FieldErrors<CreateOrEditNotificationTargetFormOutput>): NotificationSection[] {
    const sections: NotificationSection[] = [];

    if (errors.senderEmailAccountId || errors.customAddresses) {
        sections.push("email");
    }
    if (errors.slackWebhookId) {
        sections.push("slack");
    }
    if (errors.discordWebhookId) {
        sections.push("discord");
    }

    return sections;
}

function buildEmailOptions(emails: SettingEmail[]) {
    return emails.map(item => ({
        value: { id: item.id, name: item.name },
        label: item.name,
    }));
}

function buildImOptions(items: SettingImService[], kind: EImServiceKind) {
    return items
        .filter(item => item.kind === kind)
        .map(item => ({
            value: { id: item.id, name: item.name },
            label: item.name,
        }));
}

export function CreateOrEditNotificationTargetForm({
    isPending,
    onSubmit,
    onHasChanges,
    savedVersion = 0,
    initialValues,
    showAvailableInProjects = true,
    readOnly = false,
    onClose,
}: Props) {
    const isReadOnly = readOnly;
    const [openSections, setOpenSections] = useState<NotificationSection[]>(initialOpenSections);

    const emailQuery = EmailQueries.useFindManyPaginated({ pagination: selectPagination });
    const imServiceQuery = ImServiceQueries.useFindManyPaginated({ pagination: selectPagination });
    const {
        handleSubmit,
        control,
        getValues,
        reset,
        formState: { errors, isDirty },
    } = useForm<CreateOrEditNotificationTargetFormInput, unknown, CreateOrEditNotificationTargetFormOutput>({
        defaultValues: {
            name: initialValues?.name ?? "",
            emailEnabled: initialValues?.emailEnabled ?? false,
            emailUseDefault: initialValues?.emailUseDefault ?? true,
            senderEmailAccountId: initialValues?.senderEmailAccountId ?? "",
            notifyAdmins: initialValues?.notifyAdmins ?? false,
            notifyProjectOwners: initialValues?.notifyProjectOwners ?? false,
            notifyProjectMembers: initialValues?.notifyProjectMembers ?? false,
            customAddresses: initialValues?.customAddresses ?? "",
            slackEnabled: initialValues?.slackEnabled ?? false,
            slackUseDefault: initialValues?.slackUseDefault ?? true,
            slackWebhookId: initialValues?.slackWebhookId ?? "",
            discordEnabled: initialValues?.discordEnabled ?? false,
            discordUseDefault: initialValues?.discordUseDefault ?? true,
            discordWebhookId: initialValues?.discordWebhookId ?? "",
            minSendInterval: initialValues?.minSendInterval ?? "3m",
            inheritable: initialValues?.inheritable ?? false,
            default: initialValues?.default ?? false,
        },
        resolver: zodResolver(CreateOrEditNotificationTargetFormSchema),
        mode: "onSubmit",
    });

    useEffect(() => {
        if (savedVersion === 0) {
            return;
        }

        reset(getValues());
        onHasChanges?.(false);
    }, [getValues, onHasChanges, reset, savedVersion]);

    useEffect(() => {
        onHasChanges?.(isReadOnly ? false : isDirty);
    }, [isDirty, onHasChanges, isReadOnly]);

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({ name: "name", control });
    const { field: emailEnabled } = useController({ name: "emailEnabled", control });
    const { field: emailUseDefault } = useController({ name: "emailUseDefault", control });
    const {
        field: senderEmailAccountId,
        fieldState: { invalid: isSenderEmailAccountInvalid },
    } = useController({ name: "senderEmailAccountId", control });
    const { field: notifyAdmins } = useController({ name: "notifyAdmins", control });
    const { field: notifyProjectOwners } = useController({ name: "notifyProjectOwners", control });
    const { field: notifyProjectMembers } = useController({ name: "notifyProjectMembers", control });
    const {
        field: customAddresses,
        fieldState: { invalid: isCustomAddressesInvalid },
    } = useController({ name: "customAddresses", control });
    const { field: slackEnabled } = useController({ name: "slackEnabled", control });
    const { field: slackUseDefault } = useController({ name: "slackUseDefault", control });
    const {
        field: slackWebhookId,
        fieldState: { invalid: isSlackWebhookInvalid },
    } = useController({ name: "slackWebhookId", control });
    const { field: discordEnabled } = useController({ name: "discordEnabled", control });
    const { field: discordUseDefault } = useController({ name: "discordUseDefault", control });
    const {
        field: discordWebhookId,
        fieldState: { invalid: isDiscordWebhookInvalid },
    } = useController({ name: "discordWebhookId", control });
    const {
        field: minSendInterval,
        fieldState: { invalid: isMinSendIntervalInvalid },
    } = useController({ name: "minSendInterval", control });
    const { field: inheritable } = useController({ name: "inheritable", control });
    const { field: defaultField } = useController({ name: "default", control });

    const isEmailEnabled = emailEnabled.value;
    const isEmailUseDefault = emailUseDefault.value;
    const isSlackEnabled = slackEnabled.value;
    const isSlackUseDefault = slackUseDefault.value;
    const isDiscordEnabled = discordEnabled.value;
    const isDiscordUseDefault = discordUseDefault.value;

    const emailOptions = useMemo(() => buildEmailOptions(emailQuery.data?.data ?? []), [emailQuery.data?.data]);
    const slackOptions = useMemo(
        () => buildImOptions(imServiceQuery.data?.data ?? [], EImServiceKind.Slack),
        [imServiceQuery.data?.data],
    );
    const discordOptions = useMemo(
        () => buildImOptions(imServiceQuery.data?.data ?? [], EImServiceKind.Discord),
        [imServiceQuery.data?.data],
    );

    useEffect(() => {
        const option = emailOptions[0];
        if (
            isEmailEnabled &&
            !isEmailUseDefault &&
            !senderEmailAccountId.value &&
            emailOptions.length === 1 &&
            option
        ) {
            senderEmailAccountId.onChange(option.value.id);
        }
    }, [emailOptions, isEmailEnabled, isEmailUseDefault, senderEmailAccountId]);

    useEffect(() => {
        const option = slackOptions[0];
        if (isSlackEnabled && !isSlackUseDefault && !slackWebhookId.value && slackOptions.length === 1 && option) {
            slackWebhookId.onChange(option.value.id);
        }
    }, [isSlackEnabled, isSlackUseDefault, slackOptions, slackWebhookId]);

    useEffect(() => {
        const option = discordOptions[0];
        if (
            isDiscordEnabled &&
            !isDiscordUseDefault &&
            !discordWebhookId.value &&
            discordOptions.length === 1 &&
            option
        ) {
            discordWebhookId.onChange(option.value.id);
        }
    }, [discordOptions, discordWebhookId, isDiscordEnabled, isDiscordUseDefault]);

    function onValid(values: CreateOrEditNotificationTargetFormOutput) {
        if (isReadOnly) {
            return;
        }

        onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<CreateOrEditNotificationTargetFormOutput>) {
        const invalidSections = getInvalidSections(_errors);
        if (invalidSections.length > 0) {
            setOpenSections(current => mergeOpenSections(current, invalidSections));
        }
        console.error(_errors);
    }

    return (
        <form
            onSubmit={event => {
                event.preventDefault();
                void handleSubmit(onValid, onInvalid)(event);
            }}
            className="min-h-0 flex flex-1 flex-col"
        >
            <div className="">
                {readOnly && <PermissionReadonlyNotice />}
                <fieldset
                    disabled={isReadOnly}
                    className={`flex flex-col gap-6 border-0 p-0 m-0 min-w-0 ${SETTINGS_FORM_FIELD_CONTROL_MAX_WIDTH_CLASS}`}
                >
                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label="Name" />}
                    >
                        <FieldGroup>
                            <Field>
                                <Input
                                    {...name}
                                    aria-invalid={isNameInvalid}
                                />
                                <FieldError errors={[errors.name]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <Accordion
                        type="multiple"
                        value={openSections}
                        onValueChange={value => {
                            setOpenSections(value as NotificationSection[]);
                        }}
                        className="w-full flex flex-col gap-4"
                    >
                        <AccordionItem
                            value="email"
                            className="border-none"
                        >
                            <AccordionTrigger className="px-3 py-2 [&>svg]:rotate-90 [&[data-state=open]>svg]:rotate-0 bg-accent">
                                Email Notification
                            </AccordionTrigger>
                            <AccordionContent className="pt-4 pb-0 pl-4">
                                <div className="flex flex-col gap-6">
                                    <InfoBlock
                                        titleWidth={220}
                                        title={<LabelWithInfo label="Enabled" />}
                                    >
                                        <Checkbox
                                            checked={isEmailEnabled}
                                            onCheckedChange={checked => {
                                                emailEnabled.onChange(Boolean(checked));
                                            }}
                                        />
                                    </InfoBlock>
                                    {isEmailEnabled && (
                                        <>
                                            <InfoBlock
                                                titleWidth={220}
                                                title={<LabelWithInfo label="Use Default Email Account" />}
                                            >
                                                <Checkbox
                                                    checked={isEmailUseDefault}
                                                    onCheckedChange={checked => {
                                                        emailUseDefault.onChange(Boolean(checked));
                                                    }}
                                                />
                                            </InfoBlock>
                                            {!isEmailUseDefault && (
                                                <InfoBlock
                                                    titleWidth={220}
                                                    title={<LabelWithInfo label="Sender Email Account" />}
                                                >
                                                    <FieldGroup>
                                                        <Field>
                                                            <Combobox
                                                                options={emailOptions}
                                                                value={senderEmailAccountId.value}
                                                                onChange={value => {
                                                                    senderEmailAccountId.onChange(value ?? "");
                                                                }}
                                                                placeholder="None"
                                                                emptyText="No email accounts available"
                                                                searchable={false}
                                                                loading={emailQuery.isLoading}
                                                                onRefresh={() => {
                                                                    void emailQuery.refetch();
                                                                }}
                                                                isRefreshing={emailQuery.isRefetching}
                                                                aria-invalid={isSenderEmailAccountInvalid}
                                                            />
                                                            <AppLink.Basic
                                                                className="text-xs text-link"
                                                                to={ROUTE.settings.emailAccounts.$route}
                                                                target="_blank"
                                                                rel="noopener noreferrer"
                                                                ignorePrevPath
                                                            >
                                                                Configure Email Accounts
                                                            </AppLink.Basic>
                                                            <FieldError errors={[errors.senderEmailAccountId]} />
                                                        </Field>
                                                    </FieldGroup>
                                                </InfoBlock>
                                            )}
                                            <InfoBlock
                                                titleWidth={220}
                                                title={<LabelWithInfo label="Notify Admins" />}
                                            >
                                                <Checkbox
                                                    checked={notifyAdmins.value}
                                                    onCheckedChange={checked => {
                                                        notifyAdmins.onChange(Boolean(checked));
                                                    }}
                                                />
                                            </InfoBlock>
                                            <InfoBlock
                                                titleWidth={220}
                                                title={<LabelWithInfo label="Notify Project Owners" />}
                                            >
                                                <Checkbox
                                                    checked={notifyProjectOwners.value}
                                                    onCheckedChange={checked => {
                                                        notifyProjectOwners.onChange(Boolean(checked));
                                                    }}
                                                />
                                            </InfoBlock>
                                            <InfoBlock
                                                titleWidth={220}
                                                title={<LabelWithInfo label="Notify Project Members" />}
                                            >
                                                <Checkbox
                                                    checked={notifyProjectMembers.value}
                                                    onCheckedChange={checked => {
                                                        notifyProjectMembers.onChange(Boolean(checked));
                                                    }}
                                                />
                                            </InfoBlock>
                                            <InfoBlock
                                                titleWidth={220}
                                                title={<LabelWithInfo label="Custom Addresses" />}
                                            >
                                                <FieldGroup>
                                                    <Field>
                                                        <Input
                                                            {...customAddresses}
                                                            placeholder="email1, email2"
                                                            aria-invalid={isCustomAddressesInvalid}
                                                        />
                                                        <FieldError errors={[errors.customAddresses]} />
                                                    </Field>
                                                </FieldGroup>
                                            </InfoBlock>
                                        </>
                                    )}
                                </div>
                            </AccordionContent>
                        </AccordionItem>

                        <AccordionItem
                            value="slack"
                            className="border-none"
                        >
                            <AccordionTrigger className="px-3 py-2 [&>svg]:rotate-90 [&[data-state=open]>svg]:rotate-0 bg-accent">
                                Slack Notification
                            </AccordionTrigger>
                            <AccordionContent className="pt-4 pb-0 pl-4">
                                <div className="flex flex-col gap-6">
                                    <InfoBlock
                                        titleWidth={220}
                                        title={<LabelWithInfo label="Enabled" />}
                                    >
                                        <Checkbox
                                            checked={isSlackEnabled}
                                            onCheckedChange={checked => {
                                                slackEnabled.onChange(Boolean(checked));
                                            }}
                                        />
                                    </InfoBlock>
                                    {isSlackEnabled && (
                                        <>
                                            <InfoBlock
                                                titleWidth={220}
                                                title={<LabelWithInfo label="Use Default Slack Webhook" />}
                                            >
                                                <Checkbox
                                                    checked={isSlackUseDefault}
                                                    onCheckedChange={checked => {
                                                        slackUseDefault.onChange(Boolean(checked));
                                                    }}
                                                />
                                            </InfoBlock>
                                            {!isSlackUseDefault && (
                                                <InfoBlock
                                                    titleWidth={220}
                                                    title={<LabelWithInfo label="Webhook" />}
                                                >
                                                    <FieldGroup>
                                                        <Field>
                                                            <Combobox
                                                                options={slackOptions}
                                                                value={slackWebhookId.value}
                                                                onChange={value => {
                                                                    slackWebhookId.onChange(value ?? "");
                                                                }}
                                                                placeholder="None"
                                                                emptyText="No Slack webhooks available"
                                                                searchable={false}
                                                                loading={imServiceQuery.isLoading}
                                                                onRefresh={() => {
                                                                    void imServiceQuery.refetch();
                                                                }}
                                                                isRefreshing={imServiceQuery.isRefetching}
                                                                aria-invalid={isSlackWebhookInvalid}
                                                            />
                                                            <AppLink.Basic
                                                                className="text-xs text-link"
                                                                to={ROUTE.settings.imPlatforms.$route}
                                                                target="_blank"
                                                                rel="noopener noreferrer"
                                                                ignorePrevPath
                                                            >
                                                                Configure IM Platforms
                                                            </AppLink.Basic>
                                                            <FieldError errors={[errors.slackWebhookId]} />
                                                        </Field>
                                                    </FieldGroup>
                                                </InfoBlock>
                                            )}
                                        </>
                                    )}
                                </div>
                            </AccordionContent>
                        </AccordionItem>

                        <AccordionItem
                            value="discord"
                            className="border-none"
                        >
                            <AccordionTrigger className="px-3 py-2 [&>svg]:rotate-90 [&[data-state=open]>svg]:rotate-0 bg-accent">
                                Discord Notification
                            </AccordionTrigger>
                            <AccordionContent className="pt-4 pb-0 pl-4">
                                <div className="flex flex-col gap-6">
                                    <InfoBlock
                                        titleWidth={220}
                                        title={<LabelWithInfo label="Enabled" />}
                                    >
                                        <Checkbox
                                            checked={isDiscordEnabled}
                                            onCheckedChange={checked => {
                                                discordEnabled.onChange(Boolean(checked));
                                            }}
                                        />
                                    </InfoBlock>
                                    {isDiscordEnabled && (
                                        <>
                                            <InfoBlock
                                                titleWidth={220}
                                                title={<LabelWithInfo label="Use Default Discord Webhook" />}
                                            >
                                                <Checkbox
                                                    checked={isDiscordUseDefault}
                                                    onCheckedChange={checked => {
                                                        discordUseDefault.onChange(Boolean(checked));
                                                    }}
                                                />
                                            </InfoBlock>
                                            {!isDiscordUseDefault && (
                                                <InfoBlock
                                                    titleWidth={220}
                                                    title={<LabelWithInfo label="Webhook" />}
                                                >
                                                    <FieldGroup>
                                                        <Field>
                                                            <Combobox
                                                                options={discordOptions}
                                                                value={discordWebhookId.value}
                                                                onChange={value => {
                                                                    discordWebhookId.onChange(value ?? "");
                                                                }}
                                                                placeholder="None"
                                                                emptyText="No Discord webhooks available"
                                                                searchable={false}
                                                                loading={imServiceQuery.isLoading}
                                                                onRefresh={() => {
                                                                    void imServiceQuery.refetch();
                                                                }}
                                                                isRefreshing={imServiceQuery.isRefetching}
                                                                aria-invalid={isDiscordWebhookInvalid}
                                                            />
                                                            <AppLink.Basic
                                                                className="text-xs text-link"
                                                                to={ROUTE.settings.imPlatforms.$route}
                                                                target="_blank"
                                                                rel="noopener noreferrer"
                                                                ignorePrevPath
                                                            >
                                                                Configure IM Platforms
                                                            </AppLink.Basic>
                                                            <FieldError errors={[errors.discordWebhookId]} />
                                                        </Field>
                                                    </FieldGroup>
                                                </InfoBlock>
                                            )}
                                        </>
                                    )}
                                </div>
                            </AccordionContent>
                        </AccordionItem>
                    </Accordion>
                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label="Override Min Send Interval" />}
                    >
                        <FieldGroup>
                            <Field>
                                <Input
                                    {...minSendInterval}
                                    placeholder="3m"
                                    aria-invalid={isMinSendIntervalInvalid}
                                />
                                <FieldError errors={[errors.minSendInterval]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    {showAvailableInProjects && (
                        <InfoBlock
                            titleWidth={220}
                            title={<LabelWithInfo label="Available in Projects" />}
                        >
                            <div className="flex items-center gap-3">
                                <Checkbox
                                    disabled={isReadOnly}
                                    checked={inheritable.value}
                                    onCheckedChange={checked => {
                                        inheritable.onChange(Boolean(checked));
                                    }}
                                />
                            </div>
                        </InfoBlock>
                    )}

                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label="Default" />}
                    >
                        <Checkbox
                            checked={defaultField.value}
                            onCheckedChange={checked => {
                                defaultField.onChange(Boolean(checked));
                            }}
                        />
                    </InfoBlock>
                </fieldset>
            </div>
            {!isReadOnly && (
                <FormActionBar>
                    <SettingsFormCancelAction
                        onCancel={onClose}
                        disabled={isPending}
                    />
                    <Button
                        type="submit"
                        isLoading={isPending}
                        className="min-w-[100px]"
                    >
                        Save
                    </Button>
                </FormActionBar>
            )}
            {isReadOnly && (
                <FormActionBar>
                    <Button
                        type="button"
                        onClick={onClose}
                        className="min-w-[100px]"
                    >
                        Close
                    </Button>
                </FormActionBar>
            )}
        </form>
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: CreateOrEditNotificationTargetFormOutput) => void;
    onHasChanges?: (dirty: boolean) => void;
    savedVersion?: number;
    initialValues?: Partial<CreateOrEditNotificationTargetFormInput>;
    showAvailableInProjects?: boolean;
    readOnly?: boolean;
    onClose?: () => void;
}
