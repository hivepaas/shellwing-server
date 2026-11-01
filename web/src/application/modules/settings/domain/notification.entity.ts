import type { SettingsBaseEntity } from "./settings.base.entity";

export interface NotificationViaEmail {
    enabled: boolean;
    useDefault: boolean;
    sender?: SettingsBaseEntity;
    toProjectMembers: boolean;
    toProjectOwners: boolean;
    toAllAdmins: boolean;
    toAddresses: string[];
}

export interface NotificationViaSlack {
    enabled: boolean;
    useDefault: boolean;
    webhook?: SettingsBaseEntity;
}

export interface NotificationViaDiscord {
    enabled: boolean;
    useDefault: boolean;
    webhook?: SettingsBaseEntity;
}

export interface SettingNotification extends SettingsBaseEntity {
    viaEmail?: NotificationViaEmail;
    viaSlack?: NotificationViaSlack;
    viaDiscord?: NotificationViaDiscord;
    minSendInterval: string; // duration in string format from BE
    inherited?: boolean;
}
