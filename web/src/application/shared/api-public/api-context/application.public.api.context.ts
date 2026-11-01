import { createContext } from "react";

import { UsersPublicApi, UsersPublicApiValidator } from "@application/shared/api-public/services";

function createApplicationPublicApi() {
    return {
        users: new UsersPublicApi(new UsersPublicApiValidator()),
    };
}

export const ApplicationPublicApiContext = createContext({
    api: createApplicationPublicApi(),
});
