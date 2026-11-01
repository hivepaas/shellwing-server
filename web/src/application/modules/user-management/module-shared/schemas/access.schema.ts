import { z } from "zod";

const AccessActionsSchema = z.object({
    read: z.boolean(),
    execute: z.boolean(),
    write: z.boolean(),
    delete: z.boolean(),
});

export const AccessSchema = z.object({
    id: z.string(),
    name: z.string(),
    access: AccessActionsSchema,
});
