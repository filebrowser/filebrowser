const BASE_URL = "http://localhost:8080";

export async function getXauthToken() {
  try {
    const res = await fetch(`${BASE_URL}/api/login`, {
      method: "POST",
      body: JSON.stringify({
        username: "admin",
        password: "admin",
      }),
    });
    return await res.text();
  } catch (error) {
    console.error("Error requesting acces token:", error);
  }
}

export const filesToDelete = [];

export async function deleteFile(filename) {
  try {
    const res = await fetch(`${BASE_URL}/api/resources/${filename}`, {
      method: "DELETE",
      headers: {
        "X-Auth": await getXauthToken(),
      },
    });
    if (res.status == 200) {
      //remove the deleted file from filesToDelete array
      const fileIndex = filesToDelete.findIndex((file) => file == filename);
      filesToDelete.splice(fileIndex, 1);
    }
  } catch (error) {
    console.error("Error deleting file:", error);
  }
}

export async function createFile(filename) {
  try {
    await fetch(`${BASE_URL}/api/resources/${filename}`, {
      method: "POST",
      headers: {
        "X-Auth": await getXauthToken(),
      },
    });
  } catch (error) {
    console.error("Error creating file:", error);
  }
}

export const swapFileOnRename = async (oldfileName, newfileName) => {
  const fileToSwapIndex = filesToDelete.findIndex(
    (file) => file == oldfileName
  );
  filesToDelete[fileToSwapIndex] = newfileName;
};

export async function cleanUpTempFiles() {
  for (let i = 0; i < filesToDelete.length; i++) {
    await deleteFile(filesToDelete[i]);
  }
}
