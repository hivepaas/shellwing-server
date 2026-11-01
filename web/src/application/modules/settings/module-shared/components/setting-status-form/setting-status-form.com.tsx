import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm, useFormState } from "react-hook-form";
import { useUpdateEffect } from "react-use";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ESettingStatus } from "@application/shared/enums";

import {
    Button,
    Checkbox,
    DialogActionFooter,
    DialogBody,
    Field,
    FieldError,
    FieldGroup,
    Tabs,
    TabsList,
    TabsTrigger,
} from "@/components/ui";
import { DateTimePicker } from "@/components/ui/date-time-picker";

import { PermissionReadonlyNotice } from "../permission-readonly-notice.com";

import {
    type SettingStatusFormInput,
    type SettingStatusFormOutput,
    SettingStatusFormSchema,
} from "./setting-status-form.schema";

const statusMap = {
    [ESettingStatus.Active]: "Active",
    [ESettingStatus.Disabled]: "Disabled",
} as const;

/** The status, expiration, availability and default of a setting, as its status dialog changes them. */
export function SettingStatusForm({
    isPending,
    onSubmit,
    initialValues,
    onHasChanges,
    showAvailableInProjects = true,
    readOnly = false,
    onClose,
}: Props) {
    const {
        handleSubmit,
        control,
        formState: { errors },
    } = useForm<SettingStatusFormInput, unknown, SettingStatusFormOutput>({
        defaultValues: {
            status: initialValues?.status === ESettingStatus.Disabled ? ESettingStatus.Disabled : ESettingStatus.Active,
            expireAt: initialValues?.expireAt ?? undefined,
            inheritable: initialValues?.inheritable ?? false,
            default: initialValues?.default ?? false,
        },
        resolver: zodResolver(SettingStatusFormSchema),
        mode: "onSubmit",
    });

    const { isDirty } = useFormState({ control });

    useUpdateEffect(() => {
        onHasChanges?.(readOnly ? false : isDirty);
    }, [isDirty]);

    const { field: status } = useController({ name: "status", control });
    const {
        field: expireAt,
        fieldState: { invalid: isExpireAtInvalid },
    } = useController({ name: "expireAt", control });
    const { field: inheritable } = useController({ name: "inheritable", control });
    const { field: defaultField } = useController({ name: "default", control });

    function onValid(values: SettingStatusFormOutput) {
        if (readOnly) {
            return;
        }

        onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<SettingStatusFormOutput>) {
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
            <DialogBody className="flex flex-col gap-4">
                {readOnly && <PermissionReadonlyNotice />}
                <fieldset
                    disabled={readOnly}
                    className="border-0 p-0 m-0 min-w-0"
                >
                    <FieldGroup>
                        <Field>
                            <InfoBlock
                                title="Status"
                                titleWidth={160}
                            >
                                <Tabs
                                    value={status.value}
                                    onValueChange={value => {
                                        status.onChange(value);
                                    }}
                                >
                                    <TabsList>
                                        {Object.entries(statusMap).map(([value, label]) => (
                                            <TabsTrigger
                                                key={value}
                                                value={value}
                                            >
                                                {label}
                                            </TabsTrigger>
                                        ))}
                                    </TabsList>
                                </Tabs>
                            </InfoBlock>
                        </Field>

                        <Field>
                            <InfoBlock
                                title="Access expiration"
                                titleWidth={160}
                            >
                                <DateTimePicker
                                    value={expireAt.value ?? undefined}
                                    onChange={date => {
                                        expireAt.onChange(date ?? null);
                                    }}
                                    displayFormat={{ hour24: "yyyy-MM-dd HH:mm:ss" }}
                                    granularity="second"
                                    showClearButton
                                    aria-invalid={isExpireAtInvalid}
                                />
                                <FieldError errors={[errors.expireAt]} />
                            </InfoBlock>
                        </Field>

                        {showAvailableInProjects && (
                            <Field>
                                <InfoBlock
                                    title={<LabelWithInfo label="Available in Projects" />}
                                    titleWidth={160}
                                >
                                    <div className="flex items-center gap-3">
                                        <Checkbox
                                            disabled={readOnly}
                                            checked={inheritable.value}
                                            onCheckedChange={checked => {
                                                inheritable.onChange(Boolean(checked));
                                            }}
                                        />
                                    </div>
                                </InfoBlock>
                            </Field>
                        )}

                        <Field>
                            <InfoBlock
                                title={<LabelWithInfo label="Default" />}
                                titleWidth={160}
                            >
                                <Checkbox
                                    checked={defaultField.value}
                                    onCheckedChange={checked => {
                                        defaultField.onChange(Boolean(checked));
                                    }}
                                />
                            </InfoBlock>
                        </Field>
                    </FieldGroup>
                </fieldset>
            </DialogBody>
            {!readOnly && (
                <DialogActionFooter>
                    <div className="flex justify-end">
                        <Button
                            type="submit"
                            isLoading={isPending}
                            className="min-w-[100px]"
                        >
                            Save
                        </Button>
                    </div>
                </DialogActionFooter>
            )}
            {readOnly && (
                <DialogActionFooter>
                    <div className="flex justify-end">
                        <Button
                            type="button"
                            onClick={onClose}
                            className="min-w-[100px]"
                        >
                            Close
                        </Button>
                    </div>
                </DialogActionFooter>
            )}
        </form>
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: SettingStatusFormOutput) => void;
    initialValues?: Partial<SettingStatusFormInput>;
    onHasChanges?: (dirty: boolean) => void;
    showAvailableInProjects?: boolean;
    readOnly?: boolean;
    onClose?: () => void;
}
