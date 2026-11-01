import { createContext } from "react";

import {
    AuditLogsApi,
    AuditLogsApiValidator,
    SystemTaskLogsWsApi,
    SystemTasksApi,
    SystemTasksApiValidator,
} from "../services";

function createApi() {
    return {
        operations: {
            tasks: new SystemTasksApi(new SystemTasksApiValidator()),
            taskLogs: {
                $: new SystemTaskLogsWsApi(),
            },
            auditLogs: new AuditLogsApi(new AuditLogsApiValidator()),
        },
    };
}

export const OperationsApiContext = createContext({
    api: createApi(),
});
