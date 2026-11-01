import { useState } from "react";

import { toast } from "sonner";
import { CloudStorageCommands } from "~/settings/data/commands";
import { CloudStorageQueries } from "~/settings/data/queries";
import { CreateOrEditCloudStorageForm } from "~/settings/module-shared/components/cloud-storage-form";
import type {
    CreateOrEditCloudStorageFormInput,
    CreateOrEditCloudStorageFormOutput,
} from "~/settings/module-shared/components/cloud-storage-form";
import { SettingsFormRouteHeader } from "~/settings/module-shared/components/settings-form-route-header";

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { ECloudStorageKind } from "@application/shared/enums";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalModule } from "@application/shared/permissions";

type CloudStorageFormRouteMode = "create" | "edit";

export function CloudStorageFormRoute({ mode, cloudStorageId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const [saveRevision, setSaveRevision] = useState(0);
    const [testStatus, setTestStatus] = useState<"idle" | "succeeded" | "failed">("idle");
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Settings });
    const { navigate } = useAppNavigate();

    const isEditMode = mode === "edit";
    const detailId = isEditMode ? (cloudStorageId ?? "") : "";

    function navigateToList() {
        navigate.modules(ROUTE.settings.cloudStorages.$route, { ignorePrevPath: true });
    }

    function markSaved() {
        setHasChanges(false);
        setSaveRevision(revision => revision + 1);
        navigateToList();
    }

    const { mutate: createCloudStorage, isPending: isCreating } = CloudStorageCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Cloud storage created successfully");
            markSaved();
        },
    });
    const { mutate: updateCloudStorage, isPending: isUpdating } = CloudStorageCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Cloud storage updated successfully");
            markSaved();
        },
    });
    const { mutate: testConnection, isPending: isTesting } = CloudStorageCommands.useTestConn({
        onSuccess: () => {
            setTestStatus("succeeded");
        },
        onError: () => {
            setTestStatus("failed");
        },
    });

    const detailQuery = CloudStorageQueries.useFindOneById({ id: detailId }, { enabled: isEditMode });
    const cloudStorage = detailQuery.data?.data;

    function createPayload(values: CreateOrEditCloudStorageFormOutput) {
        return {
            inheritable: values.inheritable,
            default: values.default,
            kind: values.kind,
            name: values.name,
            s3: {
                keyAuth: { id: values.keyAuth?.id ?? "" },
                region: values.region,
                bucket: values.bucket,
                endpoint: values.endpoint,
            },
        };
    }

    function onSubmit(values: CreateOrEditCloudStorageFormOutput) {
        const payload = createPayload(values);

        if (isEditMode && cloudStorage) {
            updateCloudStorage({ id: cloudStorage.id, payload: { ...payload, updateVer: cloudStorage.updateVer } });
            return;
        }

        createCloudStorage({ payload });
    }

    function onTestConnection(values: CreateOrEditCloudStorageFormOutput) {
        setTestStatus("idle");
        testConnection({
            payload: {
                kind: values.kind,
                name: values.name,
                s3: {
                    keyAuth: { id: values.keyAuth?.id ?? "" },
                    region: values.region,
                    bucket: values.bucket,
                    endpoint: values.endpoint,
                },
            },
        });
    }

    function handleClose() {
        if (isPending) {
            return;
        }
        if (canWrite && hasChanges && !window.confirm("Are you sure you want to close without saving changes?")) {
            return;
        }

        navigateToList();
    }

    const isPending = isCreating || isUpdating;
    const isDetailLoading = isEditMode && detailQuery.isFetching;
    const initialValues: Partial<CreateOrEditCloudStorageFormInput> | undefined = cloudStorage
        ? {
              name: cloudStorage.name,
              kind: cloudStorage.kind === ECloudStorageKind.AWSS3 ? ECloudStorageKind.AWSS3 : ECloudStorageKind.AWSS3,
              keyAuth: cloudStorage.s3.keyAuth
                  ? { id: cloudStorage.s3.keyAuth.id, name: cloudStorage.s3.keyAuth.name }
                  : null,
              region: cloudStorage.s3.region,
              bucket: cloudStorage.s3.bucket,
              endpoint: cloudStorage.s3.endpoint,
              inheritable: Boolean(cloudStorage.inheritable),
              default: cloudStorage.default ?? false,
          }
        : undefined;
    const shouldRenderForm = mode === "create" || Boolean(initialValues);
    const title = mode === "create" ? "Create Cloud Storage" : "Edit Cloud Storage";

    return (
        <div className="flex w-full flex-col">
            <SettingsFormRouteHeader title={title} />

            {isDetailLoading && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}

            {!isDetailLoading && shouldRenderForm && (
                <CreateOrEditCloudStorageForm
                    key={detailId}
                    isPending={isPending}
                    isTesting={isTesting}
                    testStatus={testStatus}
                    onSubmit={onSubmit}
                    onTestConnection={onTestConnection}
                    onHasChanges={setHasChanges}
                    savedVersion={saveRevision}
                    initialValues={initialValues}
                    showAvailableInProjects
                    readOnly={!canWrite}
                    onClose={handleClose}
                />
            )}
        </div>
    );
}

interface Props {
    mode: CloudStorageFormRouteMode;
    cloudStorageId?: string;
}
