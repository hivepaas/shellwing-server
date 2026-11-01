export const SystemTaskType = {
    Dummy: "task:dummy",
    SchedJobExec: "task:sched-job-exec",
} as const;

export type SystemTaskType = (typeof SystemTaskType)[keyof typeof SystemTaskType];

export const SystemTaskStatus = {
    NotStarted: "not-started",
    InProgress: "in-progress",
    Canceled: "canceled",
    Done: "done",
    Failed: "failed",
} as const;

export type SystemTaskStatus = (typeof SystemTaskStatus)[keyof typeof SystemTaskStatus];

export const SystemTaskPriority = {
    Low: "low",
    Default: "default",
    Critical: "critical",
} as const;

export type SystemTaskPriority = (typeof SystemTaskPriority)[keyof typeof SystemTaskPriority];

export interface SystemTaskConfig {
    priority: SystemTaskPriority;
    maxRetry: number;
    retry: number;
    retryDelay: string;
    retryDelayIncr: string;
    retryBackoffJitter: string;
    retryDelayMax: string;
    timeout: string;
    controlDisabled: boolean;
}

/** The scheduled job a `task:sched-job-exec` task runs. */
export interface SystemTaskTargetJob {
    id: string;
    type: string;
    kind: string;
    name: string;
    status: string;
}

export interface SystemTaskScopeUser {
    id: string;
    username: string;
    email?: string;
    fullName?: string;
    photo?: string;
    role?: string;
}

export interface SystemTask {
    id: string;
    type: string;
    status: SystemTaskStatus;
    config: SystemTaskConfig;
    targetJob?: SystemTaskTargetJob;
    lastError: string;
    updateVer: number;
    scopeUser?: SystemTaskScopeUser;
    runAt: Date | null;
    retryAt: Date | null;
    startedAt: Date | null;
    endedAt: Date | null;
    createdAt: Date;
    updatedAt: Date;
}
