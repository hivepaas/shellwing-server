import { createContext } from "react";

import {
    CollectionsApi,
    CollectionsApiValidator,
    OrgGroupsApi,
    OrgGroupsApiValidator,
    OrgMembersApi,
    OrgMembersApiValidator,
    SshCaApi,
    SshCaApiValidator,
    SshKrlApi,
    SshKrlApiValidator,
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
    const orgGroupsValidator = new OrgGroupsApiValidator();
    const collectionsValidator = new CollectionsApiValidator();
    const sshCaValidator = new SshCaApiValidator();
    const sshKrlValidator = new SshKrlApiValidator();
    const sshSettingsValidator = new SshSettingsApiValidator();
    const sshPrincipalsValidator = new SshPrincipalsApiValidator();

    return {
        organization: {
            members: new OrgMembersApi(orgMembersValidator),
            groups: new OrgGroupsApi(orgGroupsValidator),
            collections: new CollectionsApi(collectionsValidator),
            ssh: {
                cas: new SshCaApi(sshCaValidator),
                krl: new SshKrlApi(sshKrlValidator),
                settings: new SshSettingsApi(sshSettingsValidator),
                principals: new SshPrincipalsApi(sshPrincipalsValidator),
            },
        },
    };
}

export const OrganizationApiContext = createContext({
    api: createApi(),
});
