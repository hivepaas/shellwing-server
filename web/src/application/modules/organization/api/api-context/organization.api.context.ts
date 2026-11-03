import { createContext } from "react";

import {
    OrgMembersApi,
    OrgMembersApiValidator,
    SshCaApi,
    SshCaApiValidator,
    SshPrincipalsApi,
    SshPrincipalsApiValidator,
    SshSettingsApi,
    SshSettingsApiValidator,
} from "~/organization/api/services";

function createApi() {
    /**
     * Organization
     */
    const orgMembersValidator = new OrgMembersApiValidator();
    const sshCaValidator = new SshCaApiValidator();
    const sshSettingsValidator = new SshSettingsApiValidator();
    const sshPrincipalsValidator = new SshPrincipalsApiValidator();

    return {
        organization: {
            members: new OrgMembersApi(orgMembersValidator),
            ssh: {
                cas: new SshCaApi(sshCaValidator),
                settings: new SshSettingsApi(sshSettingsValidator),
                principals: new SshPrincipalsApi(sshPrincipalsValidator),
            },
        },
    };
}

export const OrganizationApiContext = createContext({
    api: createApi(),
});
