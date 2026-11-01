export const MODULE_IDS = {
    Settings: "mod::settings",
    User: "mod::user",
    System: "mod::system",
} as const;

export type ResourceModuleId = (typeof MODULE_IDS)[keyof typeof MODULE_IDS];

export const MODULES = [
    {
        name: "User Management",
        id: MODULE_IDS.User,
    },
    {
        name: "Settings",
        id: MODULE_IDS.Settings,
    },
    {
        name: "System",
        id: MODULE_IDS.System,
    },
] as const;
