import { defineStore } from "pinia";
import { useFileStore } from "./file";
import { files as api } from "@/api";
import { throttle } from "lodash-es";
import buttons from "@/utils/buttons";

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

export const useUploadStore = defineStore("upload", {
  // convert to a function
  state: (): {
    id: number;
    sizes: number[];
    progress: Progress[];
    queue: UploadItem[];
    uploads: Uploads;
    speedMbyte: number;
    eta: number;
    error: Error | null;
  } => ({
    id: 0,
    sizes: [],
    progress: [],
    queue: [],
    uploads: {},
    speedMbyte: 0,
    eta: 0,
    error: null,
  }),
  getters: {
    // user and jwt getter removed, no longer needed
    getProgress: (state) => {
      if (state.progress.length === 0) {
        return 0;
      }

      const totalSize = state.sizes.reduce((a, b) => a + b, 0);

      // TODO: this looks ugly but it works with ts now
      const sum = state.progress.reduce((acc, val) => +acc + +val) as number;
      return Math.ceil((sum / totalSize) * 100);
    },
    getProgressDecimal: (state) => {
      if (state.progress.length === 0) {
        return 0;
      }

      const totalSize = state.sizes.reduce((a, b) => a + b, 0);

      // TODO: this looks ugly but it works with ts now
      const sum = state.progress.reduce((acc, val) => +acc + +val) as number;
      return ((sum / totalSize) * 100).toFixed(2);
    },
    getTotalProgressBytes: (state) => {
      if (state.progress.length === 0 || state.sizes.length === 0) {
        return "0 Bytes";
      }
      const sum = state.progress.reduce(
        (sum, p, i) =>
          (sum as number) +
          (typeof p === "number" ? p : p ? state.sizes[i] : 0),
        0
      ) as number;
      return formatSize(sum);
    },
    getTotalSize: (state) => {
      if (state.sizes.length === 0) {
        return "0 Bytes";
      }
      const totalSize = state.sizes.reduce((a, b) => a + b, 0);
      return formatSize(totalSize);
    },
    filesInUploadCount: (state) => {
      return Object.keys(state.uploads).length + state.queue.length;
    },
    filesInUpload: (state) => {
      const files = [];

      for (const index in state.uploads) {
        const upload = state.uploads[index];
        const id = upload.id;
        const type = upload.type;
        const name = upload.file.name;
        const size = state.sizes[id];
        const isDir = upload.file.isDir;
        const progress = isDir
          ? 100
          : Math.ceil(((state.progress[id] as number) / size) * 100);

        files.push({
          id,
          name,
          progress,
          type,
          isDir,
        });
      }

      return files.sort((a, b) => a.progress - b.progress);
    },
    uploadSpeed: (state) => {
      return state.speedMbyte;
    },
    getETA: (state) => state.eta,
  },
  actions: {
    // no context as first argument, use `this` instead
    setProgress({ id, loaded }: { id: number; loaded: Progress }) {
      this.progress[id] = loaded;
    },
    setError(error: Error) {
      this.error = error;
    },
    reset() {
      this.id = 0;
      this.sizes = [];
      this.progress = [];
      this.queue = [];
      this.uploads = {};
      this.speedMbyte = 0;
      this.eta = 0;
      this.error = null;
    },
    addJob(item: UploadItem) {
      this.queue.push(item);
      this.sizes[this.id] = item.file.size;
      this.id++;
    },
    moveJob() {
      const item = this.queue[0];
      this.queue.shift();
      this.uploads[item.id] = item;
    },
    removeJob(id: number) {
      delete this.uploads[id];
    },
    upload(item: UploadItem) {
      const uploadsCount = Object.keys(this.uploads).length;

      const isQueueEmpty = this.queue.length == 0;
      const isUploadsEmpty = uploadsCount == 0;

      if (isQueueEmpty && isUploadsEmpty) {
        window.addEventListener("beforeunload", beforeUnload);
        buttons.loading("upload");
      }

      this.addJob(item);
      this.processUploads();
    },
    finishUpload(item: UploadItem) {
      this.setProgress({ id: item.id, loaded: item.file.size > 0 });
      this.removeJob(item.id);
      this.processUploads();
    },
    async processUploads() {
      const uploadsCount = Object.keys(this.uploads).length;

      const isBelowLimit = uploadsCount < UPLOADS_LIMIT;
      const isQueueEmpty = this.queue.length == 0;
      const isUploadsEmpty = uploadsCount == 0;

      const isFinished = isQueueEmpty && isUploadsEmpty;
      const canProcess = isBelowLimit && !isQueueEmpty;

      if (isFinished) {
        const fileStore = useFileStore();
        window.removeEventListener("beforeunload", beforeUnload);
        buttons.success("upload");
        this.reset();
        fileStore.reload = true;
      }

      if (canProcess) {
        const item = this.queue[0];
        this.moveJob();

        if (item.file.isDir) {
          await api.post(item.path).catch(this.setError);
        } else {
          const onUpload = throttle(
            (event: ProgressEvent) =>
              this.setProgress({
                id: item.id,
                loaded: event.loaded,
              }),
            100,
            { leading: true, trailing: false }
          );

          await api
            .post(item.path, item.file.file as File, item.overwrite, onUpload)
            .catch(this.setError);
        }

        this.finishUpload(item);
      }
    },
    setUploadSpeed(value: number) {
      this.speedMbyte = value;
    },
    setETA(value: number) {
      this.eta = value;
    },
    // easily reset state using `$reset`
    clearUpload() {
      this.$reset();
    },
  },
});
