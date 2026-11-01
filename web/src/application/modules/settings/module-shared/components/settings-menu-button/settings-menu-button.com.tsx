import type { PropsWithChildren } from "react";

import { PopConfirm } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";
import type { ModuleAction } from "@application/shared/permissions";

import { Button } from "@/components/ui";

interface SettingsMenuButtonProps extends PropsWithChildren {
    action: ModuleAction;
    onClick?: () => void;
    isLoading?: boolean;
    disabled?: boolean;
}

export function SettingsMenuButton({
    action,
    onClick,
    isLoading = false,
    disabled = false,
    children,
}: SettingsMenuButtonProps) {
    return (
        <PermissionTooltipAction
            id={MODULE_IDS.Settings}
            action={action}
            triggerClassName="w-full"
        >
            {({ isDenied }) => (
                <Button
                    className="justify-start py-1.5 w-full"
                    variant="ghost"
                    onClick={onClick}
                    disabled={isDenied || disabled}
                    isLoading={isLoading}
                >
                    {children}
                </Button>
            )}
        </PermissionTooltipAction>
    );
}

interface SettingsPopConfirmButtonProps extends PropsWithChildren {
    action: ModuleAction;
    title: string;
    confirmText: string;
    cancelText: string;
    description: string;
    onConfirm: () => void;
    isLoading?: boolean;
}

export function SettingsPopConfirmButton({
    action,
    title,
    confirmText,
    cancelText,
    description,
    onConfirm,
    isLoading = false,
    children,
}: SettingsPopConfirmButtonProps) {
    const permission = useConditionalModule({ id: MODULE_IDS.Settings });

    if (!permission.hasAccess(action)) {
        return <SettingsMenuButton action={action}>{children}</SettingsMenuButton>;
    }

    return (
        <PopConfirm
            title={title}
            variant="destructive"
            confirmText={confirmText}
            cancelText={cancelText}
            description={description}
            onConfirm={onConfirm}
        >
            <Button
                className="justify-start py-1.5"
                variant="ghost"
                isLoading={isLoading}
            >
                {children}
            </Button>
        </PopConfirm>
    );
}
