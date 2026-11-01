export const AuditLogResult = {
    Allowed: "allowed",
    Denied: "denied",
} as const;

export type AuditLogResult = (typeof AuditLogResult)[keyof typeof AuditLogResult];

export const AuditLogType = {
    SecretReveal: "secret-reveal",
    UserCreate: "user-create",
    UserUpdate: "user-update",
    UserDelete: "user-delete",
    UserLogin: "user-login",
    UserLogout: "user-logout",
    SettingCreate: "setting-create",
    SettingUpdate: "setting-update",
    SettingStatusUpdate: "setting-status-update",
    SettingDelete: "setting-delete",
    TaskCancel: "task-cancel",
    HivePaaSSettingsUpdate: "hivepaas-settings-update",
    HivePaaSAction: "hivepaas-action",
} as const;

export type AuditLogType = (typeof AuditLogType)[keyof typeof AuditLogType];

export const AuditLogSource = {
    ApiGet: "api-get",
    ApiCreate: "api-create",
    ApiUpdate: "api-update",
    ApiDelete: "api-delete",
    ApiAction: "api-action",
} as const;

export type AuditLogSource = (typeof AuditLogSource)[keyof typeof AuditLogSource];

export interface AuditLogActor {
    id: string;
    type: string;
    name?: string;
    loggedName?: string;
    photo?: string;
}

export interface AuditLogResource {
    id: string;
    type: string;
    name?: string;
    loggedName?: string;
    photo?: string;
}

export interface AuditLogScopeUser {
    id: string;
    username: string;
    email?: string;
    fullName?: string;
    photo?: string;
    role?: string;
}

export interface AuditLog {
    id: string;
    type: AuditLogType;
    source?: AuditLogSource;
    /**
     * Which part of the object was written, for the types that cover several
     * endpoints - the settings tab, the form a user was edited through. It used
     * to live inside `detail`, which the listing drops once it grows past ~100
     * chars, so it went missing on exactly the entries that had the most to say.
     */
    section?: string;
    result: AuditLogResult;
    scopeUser?: AuditLogScopeUser;
    actor?: AuditLogActor;
    resource?: AuditLogResource;
    viaApiKey?: boolean;
    sessionUid?: string;
    clientIp?: string;
    remoteAddr?: string;
    userAgent?: string;
    requestId?: string;
    detail?: string;
    hasDetail?: boolean;
    createdAt: Date;
}
