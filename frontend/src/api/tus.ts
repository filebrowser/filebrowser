import * as tus from "tus-js-client";
import { baseURL, tusEndpoint, tusSettings } from "@/utils/constants";
import { useAuthStore } from "@/stores/auth";
import { useUploadStore } from "@/stores/upload";
import { removePrefix } from "@/api/utils";
import { fetchURL } from "./utils";

const RETRY_BASE_DELAY = 1000;
const RETRY_MAX_DELAY = 20000;
const SPEED_UPDATE_INTERVAL = 1000;
const ALPHA = 0.2;
const ONE_MINUS_ALPHA = 1 - ALPHA;
const RECENT_SPEEDS_LIMIT = 5;
const MB_DIVISOR = 1024 * 1024;
const CURRENT_UPLOAD_LIST: CurrentUploadList = {};

export async function upload(
  filePath: string,
  content: ApiContent = "",
  overwrite = false,
  onupload: any
) {
  if (!tusSettings) {
    // Shouldn't happen as we check for tus support before calling this function
    throw new Error("Tus.io settings are not defined");
  }

  filePath = removePrefix(filePath);
  const resourcePath = `${tusEndpoint}${filePath}?override=${overwrite}`;

  await createUpload(resourcePath);

  const authStore = useAuthStore();

  // Exit early because of typescript, tus content can't be a string
  if (content === "") {
    return false;
  }
  return new Promise<void | string>((resolve, reject) => {
    const upload = new tus.Upload(content, {
      uploadUrl: `${baseURL}${resourcePath}`,
      chunkSize: tusSettings.chunkSize,
      retryDelays: computeRetryDelays(tusSettings),
      parallelUploads: 1,
      storeFingerprintForResuming: false,
      headers: {
        "X-Auth": authStore.jwt,
      },
      onError: function (error) {
        if (CURRENT_UPLOAD_LIST[filePath].interval) {
          clearInterval(CURRENT_UPLOAD_LIST[filePath].interval);
        }
        delete CURRENT_UPLOAD_LIST[filePath];
        reject(new Error(`Upload failed: ${error.message}`));
      },
      onProgress: function (bytesUploaded) {
        const fileData = CURRENT_UPLOAD_LIST[filePath];
        fileData.currentBytesUploaded = bytesUploaded;

        if (!fileData.hasStarted) {
          fileData.hasStarted = true;
          fileData.lastProgressTimestamp = Date.now();

          fileData.interval = setInterval(() => {
            calcProgress(filePath);
          }, SPEED_UPDATE_INTERVAL);
        }
        if (typeof onupload === "function") {
          onupload({ loaded: bytesUploaded });
        }
      },
      onSuccess: function () {
        if (CURRENT_UPLOAD_LIST[filePath].interval) {
          clearInterval(CURRENT_UPLOAD_LIST[filePath].interval);
        }
        delete CURRENT_UPLOAD_LIST[filePath];
        resolve();
      },
    });
    CURRENT_UPLOAD_LIST[filePath] = {
      upload: upload,
      recentSpeeds: [],
      initialBytesUploaded: 0,
      currentBytesUploaded: 0,
      currentAverageSpeed: 0,
      lastProgressTimestamp: null,
      sumOfRecentSpeeds: 0,
      hasStarted: false,
      interval: undefined,
    };
    upload.start();
  });
}

async function createUpload(resourcePath: string) {
  const headResp = await fetchURL(resourcePath, {
    method: "POST",
  });
  if (headResp.status !== 201) {
    throw new Error(
      `Failed to create an upload: ${headResp.status} ${headResp.statusText}`
    );
  }
}

function computeRetryDelays(tusSettings: TusSettings): number[] | undefined {
  if (!tusSettings.retryCount || tusSettings.retryCount < 1) {
    // Disable retries altogether
    return undefined;
  }
  // The tus client expects our retries as an array with computed backoffs
  // E.g.: [0, 3000, 5000, 10000, 20000]
  const retryDelays = [];
  let delay = 0;

  for (let i = 0; i < tusSettings.retryCount; i++) {
    retryDelays.push(Math.min(delay, RETRY_MAX_DELAY));
    delay =
      delay === 0 ? RETRY_BASE_DELAY : Math.min(delay * 2, RETRY_MAX_DELAY);
  }

  return retryDelays;
}

export async function useTus(content: ApiContent) {
  return isTusSupported() && content instanceof Blob;
}

function isTusSupported() {
  return tus.isSupported === true;
}

function computeETA(state: ETAState, speed?: number) {
  if (state.speedMbyte === 0) {
    return Infinity;
  }
  const totalSize = state.sizes.reduce(
    (acc: number, size: number) => acc + size,
    0
  );
  const uploadedSize = state.progress.reduce(
    (acc: number, progress: Progress) => {
      if (typeof progress === "number") {
        return acc + progress;
      }
      return acc;
    },
    0
  );
  const remainingSize = totalSize - uploadedSize;
  const speedBytesPerSecond = (speed ?? state.speedMbyte) * 1024 * 1024;
  return remainingSize / speedBytesPerSecond;
}

function computeGlobalSpeedAndETA() {
  const uploadStore = useUploadStore();
  let totalSpeed = 0;
  let totalCount = 0;

  for (const filePath in CURRENT_UPLOAD_LIST) {
    totalSpeed += CURRENT_UPLOAD_LIST[filePath].currentAverageSpeed;
    totalCount++;
  }

  if (totalCount === 0) return { speed: 0, eta: Infinity };

  const averageSpeed = totalSpeed / totalCount;
  const averageETA = computeETA(uploadStore, averageSpeed);

  return { speed: averageSpeed, eta: averageETA };
}

function calcProgress(filePath: string) {
  const uploadStore = useUploadStore();
  const fileData = CURRENT_UPLOAD_LIST[filePath];

  const elapsedTime =
    (Date.now() - (fileData.lastProgressTimestamp ?? 0)) / 1000;
  const bytesSinceLastUpdate =
    fileData.currentBytesUploaded - fileData.initialBytesUploaded;
  const currentSpeed = bytesSinceLastUpdate / MB_DIVISOR / elapsedTime;

  if (fileData.recentSpeeds.length >= RECENT_SPEEDS_LIMIT) {
    fileData.sumOfRecentSpeeds -= fileData.recentSpeeds.shift() ?? 0;
  }

  fileData.recentSpeeds.push(currentSpeed);
  fileData.sumOfRecentSpeeds += currentSpeed;

  const avgRecentSpeed =
    fileData.sumOfRecentSpeeds / fileData.recentSpeeds.length;
  fileData.currentAverageSpeed =
    ALPHA * avgRecentSpeed + ONE_MINUS_ALPHA * fileData.currentAverageSpeed;

  const { speed, eta } = computeGlobalSpeedAndETA();
  uploadStore.setUploadSpeed(speed);
  uploadStore.setETA(eta);

  fileData.initialBytesUploaded = fileData.currentBytesUploaded;
  fileData.lastProgressTimestamp = Date.now();
}

export function abortAllUploads() {
  for (const filePath in CURRENT_UPLOAD_LIST) {
    if (CURRENT_UPLOAD_LIST[filePath].interval) {
      clearInterval(CURRENT_UPLOAD_LIST[filePath].interval);
    }
    if (CURRENT_UPLOAD_LIST[filePath].upload) {
      CURRENT_UPLOAD_LIST[filePath].upload.abort(true);
    }
    delete CURRENT_UPLOAD_LIST[filePath];
  }
}
