import { defineStore } from "pinia";
import { useFileStore } from "./file";
import { files as api } from "@/api";
import { throttle } from "lodash-es";
import buttons from "@/utils/buttons";
import { computed, inject, ref } from "vue";

// TODO: make this into a user setting
const UPLOADS_LIMIT = 5;

const beforeUnload = (event: Event) => {
  event.preventDefault();
  // To remove >> is deprecated
  // event.returnValue = "";
};

// Utility function to format bytes into a readable string
function formatSize(bytes: number): string {
  if (bytes === 0) return "0.00 Bytes";

  const k = 1024;
  const sizes = ["Bytes", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  // Return the rounded size with two decimal places
  return (bytes / k ** i).toFixed(2) + " " + sizes[i];
}

export const useUploadStore = defineStore("upload", () => {
  const $showError = inject<IToastError>("$showError")!;

  //
  // STATE
  //

  const allUploads = ref<Upload[]>([]);
  const activeUploads = ref<Set<Upload>>(new Set());
  const lastUpload = ref<number>(-1);
  const totalBytes = ref<number>(0);
  const sentBytes = ref<number>(0);

  //
  // GETTERS
  //

  const getProgress = computed(() => {
    return Math.ceil((sentBytes.value / totalBytes.value) * 100);
  });

  const getProgressDecimal = computed(() => {
    return ((sentBytes.value / totalBytes.value) * 100).toFixed(2);
  });

  const getTotalProgressBytes = computed(() => {
    return formatSize(sentBytes.value);
  });

  const getTotalSize = computed(() => {
    return formatSize(totalBytes.value);
  });

  //
  // ACTIONS
  //

  const reset = () => {
    allUploads.value = [];
    activeUploads.value = new Set();
    lastUpload.value = -1;
    totalBytes.value = 0;
    sentBytes.value = 0;
  };

  const nextUpload = (): Upload => {
    lastUpload.value++;

    const upload = allUploads.value[lastUpload.value];
    activeUploads.value.add(upload);

    return upload;
  };

  const hasActiveUploads = () => activeUploads.value.size > 0;

  const hasPendingUploads = () =>
    allUploads.value.length > lastUpload.value + 1;

  const upload = (
    path: string,
    name: string,
    file: File | null,
    overwrite: boolean,
    type: ResourceType
  ) => {
    if (!hasActiveUploads() && !hasPendingUploads()) {
      window.addEventListener("beforeunload", beforeUnload);
      buttons.loading("upload");
    }

    const upload: Upload = {
      path,
      name,
      file,
      overwrite,
      type,
      totalBytes: file?.size ?? 0,
      sentBytes: 0,
    };

    totalBytes.value += upload.totalBytes;
    allUploads.value.push(upload);

    processUploads();
  };

  const finishUpload = (upload: Upload) => {
    upload.sentBytes = upload.totalBytes;
    upload.file = null;

    activeUploads.value.delete(upload);
    processUploads();
  };

  const isActiveUploadsOnLimit = () => activeUploads.value.size < UPLOADS_LIMIT;

  const processUploads = async () => {
    if (!hasActiveUploads() && !hasPendingUploads()) {
      const fileStore = useFileStore();
      window.removeEventListener("beforeunload", beforeUnload);
      buttons.success("upload");
      reset();
      fileStore.reload = true;
    }

    if (isActiveUploadsOnLimit() && hasPendingUploads()) {
      const upload = nextUpload();

      if (upload.type === "dir") {
        await api.post(upload.path).catch($showError);
      } else {
        const onUpload = throttle(
          (event: ProgressEvent) => {
            const delta = event.loaded - upload.sentBytes;
            sentBytes.value += delta;

            upload.sentBytes = event.loaded;
          },
          100,
          { leading: true, trailing: false }
        );

        await api
          .post(upload.path, upload.file!, upload.overwrite, onUpload)
          .catch($showError);
      }

      finishUpload(upload);
    }
  };

  return {
    // STATE
    allUploads,
    activeUploads,
    lastUpload,
    totalBytes,
    sentBytes,

    // GETTERS
    getProgress,
    getProgressDecimal,
    getTotalProgressBytes,
    getTotalSize,

    // ACTIONS
    reset,
    upload,
    finishUpload,
    processUploads,
  };
});
