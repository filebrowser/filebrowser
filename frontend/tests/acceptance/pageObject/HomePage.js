import { format } from "util";
import { filesToDelete, swapFileOnRename } from "../../helper/file_helper.js";

export class HomePage {
  constructor() {
    this.dialogInputSelector = '//input[@class="input input--block"]';
    this.lastNavaigatedFolderSelector =
      '//div[@class="breadcrumbs"]/span[last()]/a';
    this.contentEditorSelector = '//textarea[@class="ace_text-input"]';
    this.editorContent = '//div[@class="ace_line"]';
    this.buttonSelector = `//button[@title="%s"]`;
    this.fileSelector = `//div[@aria-label="%s"]`;
    this.cardActionSelector = '//div[@class="card-action"]/button[@title="%s"]';
  }

  async createNewFolder(folderName) {
    await global.page.click(format(this.buttonSelector, "New folder"));
    await global.page.fill(this.dialogInputSelector, folderName);
    await global.page.click(format(this.cardActionSelector, "Create"));
  }

  async createFileWithContent(filename, content) {
    await global.page.click(format(this.buttonSelector, "New file"));
    await global.page.fill(this.dialogInputSelector, filename);
    await global.page.click(format(this.cardActionSelector, "Create"));
    await global.page.fill(this.contentEditorSelector, content);
    await global.page.click(format(this.buttonSelector, "Save"));
    await global.page.click(format(this.buttonSelector, "Close"));

    //saving the file info into global array to delete later
    filesToDelete.push(filename);
  }

  async renameFile(oldfileName, newfileName) {
    await global.page.click(format(this.fileSelector, oldfileName));
    await global.page.click(format(this.buttonSelector, "Rename"));
    await global.page.fill(this.dialogInputSelector, newfileName);
    await global.page.click(format(this.cardActionSelector, "Rename"));
    await swapFileOnRename(oldfileName, newfileName);
  }

  async deleteFile(filename) {
    await global.page.click(format(this.fileSelector, filename));
    await global.page.click(format(this.buttonSelector, "Delete"));
    await global.page.click(format(this.cardActionSelector, "Delete"));
  }
}
