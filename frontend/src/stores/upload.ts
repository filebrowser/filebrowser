import { defineStore } from "pinia";
import { useFileStore } from "./file";
import { files as api } from "@/api";
import { throttle } from "lodash-es";
import buttons from "@/utils/buttons";
import { computed, ref } from "vue";

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
  //
  // STATE
  //

  const id = ref<number>(0);
  const sizes = ref<number[]>([]);
  const progress = ref<number[]>([]);
  const queue = ref<UploadItem[]>([]);
  const uploads = ref<Uploads>({});
  const error = ref<Error | null>(null);

  //
  // GETTERS
  //

  const getProgress = computed(() => {
    if (progress.value.length === 0) {
      return 0;
    }

    const totalSize = sizes.value.reduce((a, b) => a + b, 0);
    const sum = progress.value.reduce((a, b) => a + b, 0);
    return Math.ceil((sum / totalSize) * 100);
  });

  const getProgressDecimal = computed(() => {
    if (progress.value.length === 0) {
      return 0;
    }

    const totalSize = sizes.value.reduce((a, b) => a + b, 0);
    const sum = progress.value.reduce((a, b) => a + b, 0);
    return ((sum / totalSize) * 100).toFixed(2);
  });

  const getTotalProgressBytes = computed(() => {
    if (progress.value.length === 0 || sizes.value.length === 0) {
      return "0 Bytes";
    }
    const sum = progress.value.reduce((a, b) => a + b, 0);
    return formatSize(sum);
  });

  const getTotalProgress = computed(() => {
    return progress.value.reduce((a, b) => a + b, 0);
  });

  const getTotalSize = computed(() => {
    if (sizes.value.length === 0) {
      return "0 Bytes";
    }
    const totalSize = sizes.value.reduce((a, b) => a + b, 0);
    return formatSize(totalSize);
  });

  const getTotalBytes = computed(() => {
    return sizes.value.reduce((a, b) => a + b, 0);
  });

  const filesInUploadCount = computed(() => {
    return Object.keys(uploads.value).length + queue.value.length;
  });

  const filesInUpload = computed(() => {
    const files = [];

    for (const index in uploads.value) {
      const upload = uploads.value[index];
      const id = upload.id;
      const type = upload.type;
      const name = upload.file.name;
      const size = sizes.value[id];
      const isDir = upload.file.isDir;
      const p = isDir ? 100 : Math.ceil((progress.value[id] / size) * 100);

      files.push({
        id,
        name,
        progress: p,
        type,
        isDir,
      });
    }

    return files.sort((a, b) => a.progress - b.progress);
  });

  //
  // ACTIONS
  //

  const setProgress = ({ id, loaded }: { id: number; loaded: number }) => {
    progress.value[id] = loaded;
  };

  const setError = (err: Error) => {
    error.value = err;
  };

  const reset = () => {
    id.value = 0;
    sizes.value = [];
    progress.value = [];
    queue.value = [];
    uploads.value = {};
    error.value = null;
  };

  const addJob = (item: UploadItem) => {
    queue.value.push(item);
    sizes.value[id.value] = item.file.size;
    id.value++;
  };

  const moveJob = () => {
    const item = queue.value[0];
    queue.value.shift();
    uploads.value[item.id] = item;
  };

  const removeJob = (id: number) => {
    delete uploads.value[id];
  };

  const upload = (item: UploadItem) => {
    const uploadsCount = Object.keys(uploads.value).length;

    const isQueueEmpty = queue.value.length == 0;
    const isUploadsEmpty = uploadsCount == 0;

    if (isQueueEmpty && isUploadsEmpty) {
      window.addEventListener("beforeunload", beforeUnload);
      buttons.loading("upload");
    }

    addJob(item);
    processUploads();
  };

  const finishUpload = (item: UploadItem) => {
    setProgress({ id: item.id, loaded: item.file.size });
    removeJob(item.id);
    processUploads();
  };

  const processUploads = async () => {
    const uploadsCount = Object.keys(uploads.value).length;

    const isBelowLimit = uploadsCount < UPLOADS_LIMIT;
    const isQueueEmpty = queue.value.length == 0;
    const isUploadsEmpty = uploadsCount == 0;

    const isFinished = isQueueEmpty && isUploadsEmpty;
    const canProcess = isBelowLimit && !isQueueEmpty;

    if (isFinished) {
      const fileStore = useFileStore();
      window.removeEventListener("beforeunload", beforeUnload);
      buttons.success("upload");
      reset();
      fileStore.reload = true;
    }

    if (canProcess) {
      const item = queue.value[0];
      moveJob();

      if (item.file.isDir) {
        await api.post(item.path).catch(setError);
      } else {
        const onUpload = throttle(
          (event: ProgressEvent) =>
            setProgress({
              id: item.id,
              loaded: event.loaded,
            }),
          100,
          { leading: true, trailing: false }
        );

        await api
          .post(item.path, item.file.file as File, item.overwrite, onUpload)
          .catch(setError);
      }

      finishUpload(item);
    }
  };

  return {
    // STATE
    id,
    sizes,
    progress,
    queue,
    uploads,
    error,

    // GETTERS
    getProgress,
    getProgressDecimal,
    getTotalProgressBytes,
    getTotalProgress,
    getTotalSize,
    getTotalBytes,
    filesInUploadCount,
    filesInUpload,

    // ACTIONS
    setProgress,
    setError,
    reset,
    addJob,
    moveJob,
    removeJob,
    upload,
    finishUpload,
    processUploads,
  };
});
