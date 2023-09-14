import * as tus from "tus-js-client";
import { baseURL, tusEndpoint, tusSettings } from "@/utils/constants";
import store from "@/store";
import { removePrefix } from "@/api/utils";
import { fetchURL } from "./utils";

const RETRY_BASE_DELAY = 1000;
const RETRY_MAX_DELAY = 20000;
const SPEED_UPDATE_INTERVAL = 1000;
const ALPHA = 0.2;
const ONE_MINUS_ALPHA = 1 - ALPHA;
const RECENT_SPEEDS_LIMIT = 5;
const MB_DIVISOR = 1024 * 1024;
const CURRENT_UPLOAD_LIST = {};

export async function upload(
  filePath,
  content = "",
  overwrite = false,
  onupload
) {
  if (!tusSettings) {
    // Shouldn't happen as we check for tus support before calling this function
    throw new Error("Tus.io settings are not defined");
  }

  filePath = removePrefix(filePath);
  let resourcePath = `${tusEndpoint}${filePath}?override=${overwrite}`;

  await createUpload(resourcePath);

  return new Promise((resolve, reject) => {
    let upload = new tus.Upload(content, {
      uploadUrl: `${baseURL}${resourcePath}`,
      chunkSize: tusSettings.chunkSize,
      retryDelays: computeRetryDelays(tusSettings),
      parallelUploads: 1,
      storeFingerprintForResuming: false,
      headers: {
        "X-Auth": store.state.jwt,
      },
      onError: function (error) {
        if (CURRENT_UPLOAD_LIST[filePath].interval) {
          clearInterval(CURRENT_UPLOAD_LIST[filePath].interval);
        }
        delete CURRENT_UPLOAD_LIST[filePath];
        reject("Upload failed: " + error);
      },
      onProgress: function (bytesUploaded) {
        let fileData = CURRENT_UPLOAD_LIST[filePath];
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
      interval: null,
    };
    upload.start();
  });
}

async function createUpload(resourcePath) {
  let headResp = await fetchURL(resourcePath, {
    method: "POST",
  });
  if (headResp.status !== 201) {
    throw new Error(
      `Failed to create an upload: ${headResp.status} ${headResp.statusText}`
    );
  }
}

function computeRetryDelays(tusSettings) {
  if (!tusSettings.retryCount || tusSettings.retryCount < 1) {
    // Disable retries altogether
    return null;
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

export async function useTus(content) {
  return isTusSupported() && content instanceof Blob;
}

function isTusSupported() {
  return tus.isSupported === true;
}

function computeETA(state) {
  if (state.speedMbyte === 0) {
    return Infinity;
  }
  const totalSize = state.sizes.reduce((acc, size) => acc + size, 0);
  const uploadedSize = state.progress.reduce(
    (acc, progress) => acc + progress,
    0
  );
  const remainingSize = totalSize - uploadedSize;
  const speedBytesPerSecond = state.speedMbyte * 1024 * 1024;
  return remainingSize / speedBytesPerSecond;
}

function computeGlobalSpeedAndETA() {
  let totalSpeed = 0;
  let totalCount = 0;

  for (let filePath in CURRENT_UPLOAD_LIST) {
    totalSpeed += CURRENT_UPLOAD_LIST[filePath].currentAverageSpeed;
    totalCount++;
  }

  if (totalCount === 0) return { speed: 0, eta: Infinity };

  const averageSpeed = totalSpeed / totalCount;
  const averageETA = computeETA(store.state.upload, averageSpeed);

  return { speed: averageSpeed, eta: averageETA };
}

function calcProgress(filePath) {
  let fileData = CURRENT_UPLOAD_LIST[filePath];

  let elapsedTime = (Date.now() - fileData.lastProgressTimestamp) / 1000;
  let bytesSinceLastUpdate =
    fileData.currentBytesUploaded - fileData.initialBytesUploaded;
  let currentSpeed = bytesSinceLastUpdate / MB_DIVISOR / elapsedTime;

  if (fileData.recentSpeeds.length >= RECENT_SPEEDS_LIMIT) {
    fileData.sumOfRecentSpeeds -= fileData.recentSpeeds.shift();
  }

  fileData.recentSpeeds.push(currentSpeed);
  fileData.sumOfRecentSpeeds += currentSpeed;

  let avgRecentSpeed =
    fileData.sumOfRecentSpeeds / fileData.recentSpeeds.length;
  fileData.currentAverageSpeed =
    ALPHA * avgRecentSpeed + ONE_MINUS_ALPHA * fileData.currentAverageSpeed;

  const { speed, eta } = computeGlobalSpeedAndETA();
  store.commit("setUploadSpeed", speed);
  store.commit("setETA", eta);

  fileData.initialBytesUploaded = fileData.currentBytesUploaded;
  fileData.lastProgressTimestamp = Date.now();
}

export function abortAllUploads() {
  for (let filePath in CURRENT_UPLOAD_LIST) {
    if (CURRENT_UPLOAD_LIST[filePath].interval) {
      clearInterval(CURRENT_UPLOAD_LIST[filePath].interval);
    }
    if (CURRENT_UPLOAD_LIST[filePath].upload) {
      CURRENT_UPLOAD_LIST[filePath].upload.abort(true);
    }
    delete CURRENT_UPLOAD_LIST[filePath];
  }
}
