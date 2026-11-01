import type { ComponentProps, PropsWithChildren } from "react";

import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction } from "@application/shared/permissions";

import { Button } from "@/components/ui";

interface Props extends PropsWithChildren {
    onClick: () => void;
    variant?: ComponentProps<typeof Button>["variant"];
}

export function SettingsCreateButton({ onClick, variant, children }: Props) {
    return (
        <PermissionTooltipAction
            id={MODULE_IDS.Settings}
            action="write"
        >
            {({ isDenied }) => (
                <Button
                    variant={variant}
                    onClick={onClick}
                    disabled={isDenied}
                >
                    {children}
                </Button>
            )}
        </PermissionTooltipAction>
    );
}
