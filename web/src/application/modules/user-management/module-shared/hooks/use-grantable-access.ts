import { useCallback } from "react";

import { useProfileContext } from "@application/shared/context";
import { EUserRole } from "@application/shared/enums";

export type AccessAction = "read" | "execute" | "write" | "delete";

/**
 * Reports which permissions the signed-in account is allowed to hand out.
 *
 * A user must not grant - or take away - more than they hold themselves; the API
 * enforces that. This mirrors the rule in the UI so the boxes out of reach are
 * visibly disabled instead of failing on save. Admins are unrestricted.
 */
export interface GrantableAccess {
    isAdmin: boolean;
    canGrantModule: (moduleId: string, action: AccessAction) => boolean;
}

export function useGrantableAccess(): GrantableAccess {
    const profile = useProfileContext(state => state.profile);
    const isAdmin = profile?.role === EUserRole.Admin;
    const moduleAccesses = profile?.moduleAccesses;

    const canGrantModule = useCallback(
        (moduleId: string, action: AccessAction) => {
            if (isAdmin) {
                return true;
            }
            const moduleAccess = (moduleAccesses ?? []).find(module => module.id === moduleId);
            return moduleAccess?.access[action] ?? false;
        },
        [isAdmin, moduleAccesses],
    );

    return { isAdmin, canGrantModule };
}
