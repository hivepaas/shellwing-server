import z from "zod";

export const CurrentPasscodeSchema = z.object({
    currentPasscode: z.string().trim().min(1, "Passcode is required"),
});

export type CurrentPasscodeSchemaInput = z.input<typeof CurrentPasscodeSchema>;
export type CurrentPasscodeSchemaOutput = z.output<typeof CurrentPasscodeSchema>;
