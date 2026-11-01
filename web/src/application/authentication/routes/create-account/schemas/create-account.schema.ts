import { z } from "zod";

export const CreateAccountSchema = z.object({
    email: z
        .string()
        .trim()
        .min(1, "Email Address is required")
        .email("Enter an email address such as you@example.com"),
    fullName: z.string().trim().min(1, "Full Name is required").max(100, "Full Name is 100 characters at most"),
});

export type CreateAccountSchemaInput = z.input<typeof CreateAccountSchema>;
export type CreateAccountSchemaOutput = z.output<typeof CreateAccountSchema>;
