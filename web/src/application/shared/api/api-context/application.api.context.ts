import { createContext } from "react";

import {
    OrgsApi,
    OrgsApiValidator,
    ProfileApi,
    ProfileApiValidator,
    SessionApi,
    SessionApiValidator,
    SupportFeedbacksApi,
    SupportFeedbacksApiValidator,
} from "@application/shared/api/services";

function createApplicationApi() {
    return {
        profile: new ProfileApi(new ProfileApiValidator()),
        session: new SessionApi(new SessionApiValidator()),
        orgs: new OrgsApi(new OrgsApiValidator()),
        support: {
            feedbacks: new SupportFeedbacksApi(new SupportFeedbacksApiValidator()),
        },
    };
}

export const ApplicationApiContext = createContext({
    api: createApplicationApi(),
});
