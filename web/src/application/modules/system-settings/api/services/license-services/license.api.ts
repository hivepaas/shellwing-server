import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    License_FindOne_Res,
    License_InstallOne_Req,
    License_InstallOne_Res,
    License_RemoveOne_Res,
} from "./license.api.contracts";
import type { LicenseApiValidator } from "./license.api.validator";

export class LicenseApi extends BaseApi {
    public constructor(private readonly validator: LicenseApiValidator) {
        super();
    }

    /**
     * The license in force, this server's ID, and how many of its users' seats are taken
     */
    async findOne(signal?: AbortSignal): Promise<Result<License_FindOne_Res, Error>> {
        return lastValueFrom(
            from(
                this.client.v1.get("/system/license", {
                    signal,
                }),
            ).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Install a license key. It is verified offline; one for another server, or past its grace, is refused.
     */
    async installOne(request: License_InstallOne_Req): Promise<Result<License_InstallOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put("/system/license", payload)).pipe(
                map(this.validator.installOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Remove the license key: the free limit is back
     */
    async removeOne(): Promise<Result<License_RemoveOne_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.delete("/system/license")).pipe(
                map(this.validator.removeOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
