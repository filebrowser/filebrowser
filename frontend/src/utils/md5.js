import sparkMD5 from 'spark-md5'

/**
 * 
 * @param {*} file 
 * @param {*default hash 2M stream file ,maxSpiece indecate how many times hash,there we need't get all file hash} maxSpiece 
 */
export async function md5Generate(file, maxSpiece = 5) {
  maxSpiece = maxSpiece > 20 ? 20 : maxSpiece < 5 ? 5 : maxSpiece;
  return new Promise((resolve, reject) => {
    var blobSlice = File.prototype.slice || File.prototype.mozSlice || File.prototype.webkitSlice,
      chunkSize = 2097152,                             // Read in chunks of 2MB
      chunks = Math.ceil(file.size / chunkSize),
      currentChunk = 0,
      spark = new sparkMD5.ArrayBuffer(),
      fileReader = new FileReader();
    if (chunks > maxSpiece) {
      chunks = maxSpiece;
    }
    fileReader.onload = function (e) {
      spark.append(e.target.result);                   // Append array buffer
      currentChunk++;

      if (currentChunk < chunks) {
        loadNext();
      } else {
        let hash = spark.end()
        console.info('computed hash', hash);  // Compute hash
        resolve(hash);
      }
    };

    fileReader.onerror = function () {
      reject('oops, something went wrong.');
    };

    function loadNext() {
      var start = currentChunk * chunkSize,
        end = ((start + chunkSize) >= file.size) ? file.size : start + chunkSize;
      fileReader.readAsArrayBuffer(blobSlice.call(file, start, end));
    }

    loadNext();
  })
}