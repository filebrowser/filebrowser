import * as tus from "tus-js-client";
import { baseURL, tusEndpoint, tusSettings } from "@/utils/constants";
import store from "@/store";
import { removePrefix } from "@/api/utils";
import { fetchURL } from "./utils";

const RETRY_BASE_DELAY = 1000;
const RETRY_MAX_DELAY = 20000;

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
        reject("Upload failed: " + error);
      },
      onProgress: function (bytesUploaded) {
        // Emulate ProgressEvent.loaded which is used by calling functions
        // loaded is specified in bytes (https://developer.mozilla.org/en-US/docs/Web/API/ProgressEvent/loaded)
        if (typeof onupload === "function") {
          onupload({ loaded: bytesUploaded });
        }
      },
      onSuccess: function () {
        resolve();
      },
    });
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
