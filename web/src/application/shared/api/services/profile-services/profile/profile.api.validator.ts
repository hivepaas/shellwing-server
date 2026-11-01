import { z } from "zod";

import type { Profile_GetProfile2FASetup_Res } from "@application/shared/api/services";

import { type ApiHttpResponse, parseApiResponse } from "@infrastructure/api";

/**
 * Get profile 2FA setup API response schema
 */
const GetProfile2FASetupSchema = z.object({
    data: z.object({
        totpToken: z.string(),
        secret: z.string(),
        qrCode: z.object({
            dataBase64: z.string(),
        }),
    }),
});

export class ProfileApiValidator {
    /**
     * Validate and transform get profile 2FA setup API response
     */
    getProfile2FASetup = (response: ApiHttpResponse): Profile_GetProfile2FASetup_Res => {
        const { data } = parseApiResponse({
            response,
            schema: GetProfile2FASetupSchema,
        });

        return {
            data: {
                totpToken: data.totpToken,
                totpQRCode: data.qrCode.dataBase64,
                secretKey: data.secret,
            },
        };
    };
}
