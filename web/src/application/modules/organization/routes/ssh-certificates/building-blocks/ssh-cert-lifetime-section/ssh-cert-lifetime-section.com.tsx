import { Button } from "@components/ui/button";
import { Field, FieldError } from "@components/ui/field";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import { toast } from "sonner";
import { SshSettingsCommands } from "~/organization/data/commands";
import {
    SSH_CERT_TTL_DEFAULT_MINUTES,
    SSH_CERT_TTL_MAX_MINUTES,
    SSH_CERT_TTL_MIN_MINUTES,
} from "~/organization/module-shared/constants";
import { formatMinutes } from "~/organization/module-shared/utils";

import { ContentBlock, InfoBlock, InputNumberWithAddon } from "@application/shared/components";

import {
    type SshCertLifetimeFormInput,
    type SshCertLifetimeFormOutput,
    SshCertLifetimeFormSchema,
} from "../../schemas";

function LifetimeForm({ orgId, certTtlMinutes }: Props) {
    const {
        control,
        handleSubmit,
        reset,
        formState: { errors, isDirty },
    } = useForm<SshCertLifetimeFormInput, unknown, SshCertLifetimeFormOutput>({
        defaultValues: { certTtlMinutes },
        resolver: zodResolver(SshCertLifetimeFormSchema),
        mode: "onSubmit",
    });

    const {
        field: certTtl,
        fieldState: { invalid: isCertTtlInvalid },
    } = useController({
        control,
        name: "certTtlMinutes",
    });

    const { mutate: updateSettings, isPending } = SshSettingsCommands.useUpdateOne({
        onSuccess: (_, request) => {
            toast.success("Certificate lifetime saved. It applies to the certificates signed from now on.");
            reset({ certTtlMinutes: request.payload.certTtlMinutes });
        },
    });

    function onValid(values: SshCertLifetimeFormOutput) {
        updateSettings({ orgId, payload: { certTtlMinutes: values.certTtlMinutes } });
    }

    function onInvalid(_errors: FieldErrors<SshCertLifetimeFormInput>) {
        // The errors are shown beside the field.
    }

    const isValidNumber = Number.isInteger(certTtl.value) && certTtl.value > 0;

    return (
        <form
            onSubmit={event => {
                event.preventDefault();
                void handleSubmit(onValid, onInvalid)(event);
            }}
        >
            <InfoBlock
                title="Lifetime"
                description={`${SSH_CERT_TTL_MIN_MINUTES} minutes to a day (${SSH_CERT_TTL_MAX_MINUTES}). ${formatMinutes(SSH_CERT_TTL_DEFAULT_MINUTES)} (${SSH_CERT_TTL_DEFAULT_MINUTES}) unless set.`}
                titleWidth={220}
            >
                <Field>
                    <div className="flex flex-wrap items-center gap-3">
                        <InputNumberWithAddon
                            value={certTtl.value}
                            onValueChange={value => {
                                certTtl.onChange(value);
                            }}
                            onBlur={certTtl.onBlur}
                            name={certTtl.name}
                            min={SSH_CERT_TTL_MIN_MINUTES}
                            max={SSH_CERT_TTL_MAX_MINUTES}
                            useGrouping={false}
                            addonRight="minutes"
                            classNameContainer="w-[220px]"
                            aria-invalid={isCertTtlInvalid}
                        />
                        <Button
                            type="submit"
                            isLoading={isPending}
                            disabled={!isDirty}
                        >
                            Save
                        </Button>
                    </div>
                    {isValidNumber && !isCertTtlInvalid && (
                        <span className="text-sm text-muted-foreground">{formatMinutes(certTtl.value)}</span>
                    )}
                    <FieldError errors={[errors.certTtlMinutes]} />
                </Field>
            </InfoBlock>
        </form>
    );
}

export function SshCertLifetimeSection({ orgId, certTtlMinutes }: Props) {
    return (
        <ContentBlock label="Certificate lifetime">
            <div className="flex flex-col gap-4">
                <p className="text-sm text-muted-foreground">
                    How long a certificate lasts once signed. Taking a person&apos;s accounts away, or disabling them,
                    revokes the certificates they hold - but a revocation only reaches a server once it has fetched the
                    list again, and this is what bounds the worst case. Keep it short.
                </p>
                {/* Saved elsewhere, or read back after saving: the form starts again from what is in force. */}
                <LifetimeForm
                    key={certTtlMinutes}
                    orgId={orgId}
                    certTtlMinutes={certTtlMinutes}
                />
            </div>
        </ContentBlock>
    );
}

interface Props {
    orgId: string;
    certTtlMinutes: number;
}
