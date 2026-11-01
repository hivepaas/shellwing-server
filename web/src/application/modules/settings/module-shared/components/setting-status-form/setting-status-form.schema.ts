import { z } from "zod";

import { ESettingStatus } from "@application/shared/enums";

export const SettingStatusFormSchema = z.object({
    status: z.enum([ESettingStatus.Active, ESettingStatus.Disabled]),
    expireAt: z.date().optional().nullable(),
    inheritable: z.boolean(),
    default: z.boolean(),
});

export type SettingStatusFormInput = z.input<typeof SettingStatusFormSchema>;
export type SettingStatusFormOutput = z.output<typeof SettingStatusFormSchema>;
