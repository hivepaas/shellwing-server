import { z } from "zod";

/** The server's limits on a collection. */
export const COLLECTION_NAME_MAX_LENGTH = 100;
export const COLLECTION_DESCRIPTION_MAX_LENGTH = 500;

export const EditCollectionFormSchema = z.object({
    name: z
        .string()
        .trim()
        .min(1, "Give the collection a name")
        .max(COLLECTION_NAME_MAX_LENGTH, `A name can be up to ${COLLECTION_NAME_MAX_LENGTH} characters`),
    description: z
        .string()
        .trim()
        .max(
            COLLECTION_DESCRIPTION_MAX_LENGTH,
            `A description can be up to ${COLLECTION_DESCRIPTION_MAX_LENGTH} characters`,
        ),
});

export type EditCollectionFormInput = z.input<typeof EditCollectionFormSchema>;
export type EditCollectionFormOutput = z.output<typeof EditCollectionFormSchema>;
