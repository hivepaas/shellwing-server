import { createContext } from "react";

import {
    CloudStorageApi,
    CloudStorageApiValidator,
    EmailApi,
    EmailApiValidator,
    ImServiceApi,
    ImServiceApiValidator,
    KeyAuthApi,
    KeyAuthApiValidator,
    NotificationsApi,
    NotificationsApiValidator,
    OAuthApi,
    OAuthApiValidator,
} from "../services";

function createApi() {
    /**
     * Settings
     */
    const notificationsValidator = new NotificationsApiValidator();
    const imServiceValidator = new ImServiceApiValidator();
    const emailValidator = new EmailApiValidator();
    const cloudStorageValidator = new CloudStorageApiValidator();
    const oauthValidator = new OAuthApiValidator();
    const keyAuthValidator = new KeyAuthApiValidator();

    return {
        settings: {
            notifications: new NotificationsApi(notificationsValidator),
            imService: new ImServiceApi(imServiceValidator),
            email: new EmailApi(emailValidator),
            cloudStorage: new CloudStorageApi(cloudStorageValidator),
            oauth: new OAuthApi(oauthValidator),
            keyAuth: new KeyAuthApi(keyAuthValidator),
        },
    };
}

export const SettingsApiContext = createContext({
    api: createApi(),
});
