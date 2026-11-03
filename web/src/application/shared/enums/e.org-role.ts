export const EOrgRole = {
    Owner: "owner",
    Admin: "admin",
    Member: "member",
} as const;

export type EOrgRole = (typeof EOrgRole)[keyof typeof EOrgRole];
