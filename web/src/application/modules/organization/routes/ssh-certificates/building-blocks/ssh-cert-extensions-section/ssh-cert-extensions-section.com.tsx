import { useState } from "react";

import { Button } from "@components/ui/button";
import { Checkbox } from "@components/ui/checkbox";
import { toast } from "sonner";
import { SshSettingsCommands } from "~/organization/data/commands";
import { SSH_CERT_EXTENSIONS } from "~/organization/module-shared/constants";

import { ContentBlock } from "@application/shared/components";

interface Props {
    orgId: string;
    certExtensions: string[];
}

function ExtensionsForm({ orgId, certExtensions }: Props) {
    const [chosen, setChosen] = useState<string[]>(certExtensions);

    const { mutate: updateSettings, isPending } = SshSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success(
                "Saved. The certificates people hold are revoked, and the next one each gets follows the change.",
            );
        },
    });

    // Compared as sets: the order is the server's own, and the boxes are shown in it.
    const isDirty = chosen.length !== certExtensions.length || chosen.some(name => !certExtensions.includes(name));

    return (
        <form
            className="flex flex-col gap-4"
            onSubmit={event => {
                event.preventDefault();
                updateSettings({ orgId, payload: { certExtensions: chosen } });
            }}
        >
            <ul className="flex flex-col gap-3">
                {SSH_CERT_EXTENSIONS.map(extension => (
                    <li
                        key={extension.name}
                        className="flex items-start gap-3"
                    >
                        <Checkbox
                            id={`ssh-ext-${extension.name}`}
                            className="mt-0.5"
                            checked={chosen.includes(extension.name)}
                            onCheckedChange={checked => {
                                setChosen(current =>
                                    checked === true
                                        ? [...current, extension.name]
                                        : current.filter(name => name !== extension.name),
                                );
                            }}
                        />
                        <label
                            htmlFor={`ssh-ext-${extension.name}`}
                            className="flex flex-col gap-0.5"
                        >
                            <span className="text-sm font-medium">{extension.label}</span>
                            <span className="text-sm text-muted-foreground">{extension.description}</span>
                            <code className="font-mono text-xs text-muted-foreground">{extension.name}</code>
                        </label>
                    </li>
                ))}
            </ul>
            <div>
                <Button
                    type="submit"
                    isLoading={isPending}
                    disabled={!isDirty}
                >
                    Save
                </Button>
            </div>
        </form>
    );
}

export function SshCertExtensionsSection({ orgId, certExtensions }: Props) {
    return (
        <ContentBlock label="What certificates may do">
            <div className="flex flex-col gap-4">
                <p className="text-sm text-muted-foreground">
                    A target server refuses anything a certificate does not allow, whatever its own configuration says,
                    so this is the narrower of the two gates. Saving a change revokes the certificates people hold -
                    each of those still allows what it allowed when it was signed - and their apps ask for new ones.
                </p>
                {/* Saved elsewhere, or read back after saving: the boxes start again from what is in force. */}
                <ExtensionsForm
                    key={certExtensions.join(",")}
                    orgId={orgId}
                    certExtensions={certExtensions}
                />
            </div>
        </ContentBlock>
    );
}
