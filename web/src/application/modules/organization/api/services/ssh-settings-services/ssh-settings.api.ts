import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type { SshSettings_UpdateOne_Req, SshSettings_UpdateOne_Res } from "./ssh-settings.api.contracts";
import type { SshSettingsApiValidator } from "./ssh-settings.api.validator";

export class SshSettingsApi extends BaseApi {
    public constructor(private readonly validator: SshSettingsApiValidator) {
        super();
    }

    /**
     * Set the organization's SSH settings: how long its certificates last
     */
    async updateOne(request: SshSettings_UpdateOne_Req): Promise<Result<SshSettings_UpdateOne_Res, Error>> {
        const { orgId, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/orgs/${orgId}/ssh/settings`, payload)).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
