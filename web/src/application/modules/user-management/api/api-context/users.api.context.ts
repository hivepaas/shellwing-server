import { createContext } from "react";

import { UserIdentityApi, UserIdentityApiValidator, UsersApi, UsersApiValidator } from "~/user-management/api/services";

function createApi() {
    /**
     * Users
     */
    const usersApiValidator = new UsersApiValidator();
    const userIdentityApiValidator = new UserIdentityApiValidator();

    return {
        users: {
            $: new UsersApi(usersApiValidator),
            identity: new UserIdentityApi(userIdentityApiValidator),
        },
    };
}

export const UsersApiContext = createContext({
    api: createApi(),
});
