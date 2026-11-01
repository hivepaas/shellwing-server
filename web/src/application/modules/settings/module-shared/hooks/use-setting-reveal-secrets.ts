import { useState } from "react";

import { isAxiosError } from "axios";
import { toast } from "sonner";

import { CAPABILITY_IDS } from "@application/shared/constants";
import { useProfileContext } from "@application/shared/context";
import { EUserRole } from "@application/shared/enums";
import { useCapability } from "@application/shared/permissions/hooks/use-capability";

import { isHttpException, isValidationException, parseApiError } from "@infrastructure/api";
import { createApiClient } from "@infrastructure/api/client";

interface ApiErrorData {
    detail?: string;
    message?: string;
    title?: string;
    error?: string;
    errors?: { message?: string }[];
}

function extractApiErrorMessage(error: unknown, fallbackMessage: string): string {
    if (typeof error === "string" && error.trim()) {
        return error.trim();
    }

    if (isAxiosError(error)) {
        const data = error.response?.data as unknown;
        if (typeof data === "string" && data.trim()) {
            return data.trim();
        }
        if (data && typeof data === "object") {
            const apiData = data as ApiErrorData;
            if (Array.isArray(apiData.errors) && apiData.errors.length > 0) {
                const firstMsg = apiData.errors[0]?.message;
                if (firstMsg?.trim()) {
                    return firstMsg.trim();
                }
            }
            if (apiData.detail?.trim()) {
                return apiData.detail.trim();
            }
            if (apiData.message?.trim()) {
                return apiData.message.trim();
            }
            if (apiData.title?.trim()) {
                return apiData.title.trim();
            }
            if (apiData.error?.trim()) {
                return apiData.error.trim();
            }
        }
    }

    try {
        const parsed = parseApiError(error);
        if (isHttpException(parsed)) {
            if (isValidationException(parsed) && parsed.errors.length > 0) {
                const firstMsg = parsed.errors[0]?.message;
                if (firstMsg?.trim()) {
                    return firstMsg.trim();
                }
            }
            if (parsed.problem.detail.trim()) {
                return parsed.problem.detail.trim();
            }
            if (parsed.problem.title.trim()) {
                return parsed.problem.title.trim();
            }
        }
        if (parsed.message.trim() && parsed.message !== "Unexpected error happened") {
            return parsed.message.trim();
        }
    } catch {
        // ignore
    }

    if (error instanceof Error && error.message.trim() && error.message !== "Unexpected error happened") {
        return error.message.trim();
    }

    return fallbackMessage;
}

export interface UseSettingRevealSecretsOptions {
    settingType: string;
    settingId?: string;
    mode?: "create" | "edit";
}

export interface UseSettingRevealSecretsResult<T> {
    canShowRevealButton: boolean;
    isDialogOpen: boolean;
    setIsDialogOpen: (open: boolean) => void;
    isRevealing: boolean;
    isRevealed: boolean;
    revealedData: T | null;
    revealRevision: number;
    handleConfirmReveal: () => Promise<void>;
}

export function useSettingRevealSecrets<T>({
    settingType,
    settingId,
    mode,
}: UseSettingRevealSecretsOptions): UseSettingRevealSecretsResult<T> {
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [isRevealing, setIsRevealing] = useState(false);
    const [isRevealed, setIsRevealed] = useState(false);
    const [revealedData, setRevealedData] = useState<T | null>(null);
    const [revealRevision, setRevealRevision] = useState(0);

    const profile = useProfileContext(state => state.profile);
    const isAdmin = profile?.role === EUserRole.Admin;
    const { hasCapability: canRevealSecrets } = useCapability(CAPABILITY_IDS.SecretReveal);

    const isModeValid = mode ? mode === "edit" : true;
    const hasValidTarget = isModeValid && Boolean(settingId);
    const canShowRevealButton = hasValidTarget && (isAdmin || canRevealSecrets);

    const handleConfirmReveal = async () => {
        if (!settingId || !canShowRevealButton || isRevealing) {
            return;
        }

        setIsRevealing(true);
        try {
            const client = createApiClient();
            const response = await client.v1.get<{ data: T }>(
                `/settings/${settingType}/${settingId}?revealSecrets=true`,
            );
            const { data } = response.data;
            if (data) {
                setRevealedData(data);
                setIsRevealed(true);
                setRevealRevision(r => r + 1);
                setIsDialogOpen(false);
                toast.success("Secrets revealed successfully");
            }
        } catch (error) {
            console.error("Failed to reveal secrets:", error);
            toast.error(extractApiErrorMessage(error, "Failed to reveal secrets"));
        } finally {
            setIsRevealing(false);
        }
    };

    return {
        canShowRevealButton,
        isDialogOpen,
        setIsDialogOpen,
        isRevealing,
        isRevealed,
        revealedData,
        revealRevision,
        handleConfirmReveal,
    };
}
