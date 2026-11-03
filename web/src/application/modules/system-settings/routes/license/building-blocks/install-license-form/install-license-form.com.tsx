import { Button } from "@components/ui/button";
import { Field, FieldError, FieldLabel } from "@components/ui/field";
import { Textarea } from "@components/ui/textarea";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import { toast } from "sonner";
import { LicenseCommands } from "~/system-settings/data/commands";

import { ContentBlock } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { type InstallLicenseFormInput, type InstallLicenseFormOutput, InstallLicenseFormSchema } from "../../schemas";

export function InstallLicenseForm({ hasKey }: Props) {
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });
    const {
        control,
        handleSubmit,
        reset,
        formState: { errors },
    } = useForm<InstallLicenseFormInput, unknown, InstallLicenseFormOutput>({
        defaultValues: { key: "" },
        resolver: zodResolver(InstallLicenseFormSchema),
        mode: "onSubmit",
    });

    const {
        field: key,
        fieldState: { invalid: isKeyInvalid },
    } = useController({
        control,
        name: "key",
    });

    const { mutate: installLicense, isPending } = LicenseCommands.useInstallOne({
        onSuccess: () => {
            toast.success("License key installed");
            reset({ key: "" });
        },
    });

    function onValid(values: InstallLicenseFormOutput) {
        if (!canWrite) {
            return;
        }

        installLicense({ payload: { key: values.key } });
    }

    function onInvalid(_errors: FieldErrors<InstallLicenseFormInput>) {
        // The errors are shown beside the field.
    }

    return (
        <ContentBlock label={hasKey ? "Replace the license key" : "Install a license key"}>
            <form
                className="flex flex-col gap-4"
                onSubmit={event => {
                    event.preventDefault();
                    void handleSubmit(onValid, onInvalid)(event);
                }}
            >
                <p className="text-sm text-muted-foreground">
                    A key is verified offline: the server never needs the portal to run. One for another server, or past
                    its grace, is refused.{hasKey ? " Installing a key replaces the one installed." : null}
                </p>
                <Field>
                    <FieldLabel htmlFor="license-key">License key</FieldLabel>
                    <Textarea
                        id="license-key"
                        {...key}
                        rows={6}
                        minRows={6}
                        maxRows={14}
                        placeholder="Paste the key as you received it"
                        className="font-mono text-xs"
                        spellCheck={false}
                        autoCapitalize="off"
                        autoCorrect="off"
                        disabled={!canWrite}
                        aria-invalid={isKeyInvalid}
                    />
                    <FieldError errors={[errors.key]} />
                </Field>
                <div className="flex justify-end">
                    <PermissionTooltipAction
                        id={MODULE_IDS.System}
                        action="write"
                    >
                        {({ isDenied }) => (
                            <Button
                                type="submit"
                                isLoading={isPending}
                                disabled={isDenied}
                            >
                                Install
                            </Button>
                        )}
                    </PermissionTooltipAction>
                </div>
            </form>
        </ContentBlock>
    );
}

interface Props {
    hasKey: boolean;
}
