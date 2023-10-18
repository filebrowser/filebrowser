class CreateFilePage {
    constructor() {
        //url
        this.url = 'http://localhost:8080'
        this.loginUrl = 'http://localhost:8080/login'
        this.fileUrl = this.url + '/files/'

        //define selectors
        this.uploadButtonSelector = '//button[@title="Upload"]/i';
        this.newFileLabelSelector = '//button[@aria-label="New file"]';
        this.writeNewFileInputSelector = '//input[@class="input input--block"]';
        this.createButtonSelector = '//button[contains(text(),"Create")]';
        this.contentBoxSelector = '//textarea[@class="ace_text-input"]';
        this.saveIconSelector = '//i[contains(text(),"save")]';
        this.closeIconSelector = '//i[contains(text(),"close")]';
    }

    async createNewFile(filename,content) {
        await page.click(this.newFileLabelSelector);
        await page.fill(this.writeNewFileInputSelector, filename);
        await page.click(this.createButtonSelector);
        // await page.getByRole('textbox').fill(content);
        await page.fill(this.contentBoxSelector,content);
        await page.click(this.saveIconSelector);
        await page.click(this.closeIconSelector);
    }
}

module.exports = CreateFilePage