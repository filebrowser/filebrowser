import * as tus from "tus-js-client";
import { tusEndpoint, tusSettings } from "@/utils/constants";
import store from "@/store";
import { removePrefix } from "@/api/utils";

const RETRY_BASE_DELAY = 1000;
const RETRY_MAX_DELAY = 20000;

export async function upload(url, content = "", overwrite = false, onupload) {
  if (!tusSettings) {
    // Shouldn't happen as we check for tus support before calling this function
    throw new Error("Tus.io settings are not defined");
  }

  return new Promise((resolve, reject) => {
    const metadata = {
      overwrite: overwrite.toString(),
      // url is URI encoded and needs to be decoded for metadata first
      destination: decodeURIComponent(removePrefix(url)),
    };
    var upload = new tus.Upload(content, {
      endpoint: tusEndpoint,
      chunkSize: tusSettings.chunkSize,
      retryDelays: computeRetryDelays(tusSettings),
      parallelUploads: tusSettings.parallelUploads || 1,
      headers: {
        "X-Auth": store.state.jwt,
        // Send the metadata with every request
        // If we used the tus client's metadata option, it would only be sent
        // with some of the requests.
        "Upload-Metadata": Object.entries(metadata)
          .map(([key, value]) => `${key} ${btoa(value)}`)
          .join(","),
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
    upload.findPreviousUploads().then(function (previousUploads) {
      if (previousUploads.length) {
        upload.resumeFromPreviousUpload(previousUploads[0]);
      }
    });
    upload.start();
  });
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
  if (!isTusSupported() || !(content instanceof Blob)) {
    return false;
  }

  // use tus if tus uploads are enabled and the content's size is larger than chunkSize
  return (
    tusSettings &&
    tusSettings.enabled === true &&
    content.size > tusSettings.chunkSize
  );
}

function isTusSupported() {
  return tus.isSupported === true;
}
