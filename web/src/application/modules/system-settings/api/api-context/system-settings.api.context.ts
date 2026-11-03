import { createContext } from "react";

import { LicenseApi, LicenseApiValidator } from "~/system-settings/api/services";

function createApi() {
    /**
     * System settings
     */
    const licenseValidator = new LicenseApiValidator();

    return {
        system: {
            license: new LicenseApi(licenseValidator),
        },
    };
}

export const SystemSettingsApiContext = createContext({
    api: createApi(),
});
