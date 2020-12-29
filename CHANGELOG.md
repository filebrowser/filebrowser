# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

## [2.11.0](https://github.com/filebrowser/filebrowser/compare/v2.10.0...v2.11.0) (2020-12-28)


### Features

* add sharing management ([#1178](https://github.com/filebrowser/filebrowser/issues/1178)) (closes [#1000](https://github.com/filebrowser/filebrowser/issues/1000)) ([677bce3](https://github.com/filebrowser/filebrowser/commit/677bce376b024d9ff38f34e74243034fe5a1ec3c))
* download shared subdirectory ([#1184](https://github.com/filebrowser/filebrowser/issues/1184)) ([fb5b28d](https://github.com/filebrowser/filebrowser/commit/fb5b28d9cbdee10d38fcd719b9fd832121be58ef))


### Bug Fixes

* check user input to prevent permission elevation ([#1196](https://github.com/filebrowser/filebrowser/issues/1196)) (closes [#1195](https://github.com/filebrowser/filebrowser/issues/1195)) ([f62806f](https://github.com/filebrowser/filebrowser/commit/f62806f6c9e9c7f392d1b747d65b8fe40b313e89))
* delete extra remove prefix ([#1186](https://github.com/filebrowser/filebrowser/issues/1186)) ([7a5298a](https://github.com/filebrowser/filebrowser/commit/7a5298a7556f7dcc52f59b8ea76d040d3ddc3d12))
* move files between different volumes (closes [#1177](https://github.com/filebrowser/filebrowser/issues/1177)) ([58835b7](https://github.com/filebrowser/filebrowser/commit/58835b7e535cc96e1c8a5d85821c1545743ca757))
* recaptcha race condition ([#1176](https://github.com/filebrowser/filebrowser/issues/1176)) ([ac3673e](https://github.com/filebrowser/filebrowser/commit/ac3673e111afac6616af9650ca07028b6c27e6cd))

## [2.10.0](https://github.com/filebrowser/filebrowser/compare/v2.9.0...v2.10.0) (2020-11-24)


### Features

* add hide dotfiles param  ([#1148](https://github.com/filebrowser/filebrowser/issues/1148)) ([10e399b](https://github.com/filebrowser/filebrowser/commit/10e399b3c3dbdcfb4465a9d4138e1da6bae0873d))
* add single click mode ([#1139](https://github.com/filebrowser/filebrowser/issues/1139)) ([e8b4e9a](https://github.com/filebrowser/filebrowser/commit/e8b4e9af46d6e99dbeb965dd9727d9ed017d52a2))
* automatically jump to the next photo when deleting while previewing ([#1143](https://github.com/filebrowser/filebrowser/issues/1143)) ([9515cee](https://github.com/filebrowser/filebrowser/commit/9515ceeb42e5ef5267400220a2082dec775e843d))
* shared folder file listing ([e119bc5](https://github.com/filebrowser/filebrowser/commit/e119bc55ea82cefcbcc0571650107dfd5d73f570))
* shared item information ([36cacdf](https://github.com/filebrowser/filebrowser/commit/36cacdf598e4e09f064c8ace0ca7a6c24b23028e))


### Bug Fixes

* empty folder in archive ([7096b3d](https://github.com/filebrowser/filebrowser/commit/7096b3dab92441981c9964e4a6175af0a255d2be))
* fix hanging when reading a named pipe file (closes [#1155](https://github.com/filebrowser/filebrowser/issues/1155)) ([586d198](https://github.com/filebrowser/filebrowser/commit/586d198d47b525eeccc6fe587573a3ad83adb4f6))
* previewer title overflow ([4e48ffc](https://github.com/filebrowser/filebrowser/commit/4e48ffc14d09dabeea12dc495144277db62b5b7d))
* resource rename action invalid path ([1ce3068](https://github.com/filebrowser/filebrowser/commit/1ce3068a99c80c153fd41359255d173bce6e79e8))

## [2.9.0](https://github.com/filebrowser/filebrowser/compare/v2.8.0...v2.9.0) (2020-10-21)


### Features

* support WKWebview custom protocol ([#1113](https://github.com/filebrowser/filebrowser/issues/1113)) ([0ac80e8](https://github.com/filebrowser/filebrowser/commit/0ac80e8387a69924284259bde448af2813d84ed1))


### Bug Fixes

* allow start from Windows explorer ([f2c4e78](https://github.com/filebrowser/filebrowser/commit/f2c4e78381610879eda5316d38a999c89df6c14a))
* file upload missing path slash ([5e27ba5](https://github.com/filebrowser/filebrowser/commit/5e27ba5c8c1be603c6ae7fec8de48e3532dea1f7))
* preview case sensitive file extension ([05bff54](https://github.com/filebrowser/filebrowser/commit/05bff54b71543fd232f1089c40504d0cbfd106be))
* search missing path slash ([2bd163d](https://github.com/filebrowser/filebrowser/commit/2bd163d92a856d65c8d4615e37898470c1edf2f4))

## [2.8.0](https://github.com/filebrowser/filebrowser/compare/v2.7.0...v2.8.0) (2020-10-05)


### Features

* add disable exec flag ([#1090](https://github.com/filebrowser/filebrowser/issues/1090)) ([97693cc](https://github.com/filebrowser/filebrowser/commit/97693cc6117ce1c956baede91de5dd48b904e175))


### Bug Fixes

* empty commands setting ([c6d4fcd](https://github.com/filebrowser/filebrowser/commit/c6d4fcd08f5f1531c2cef514dc86019e23e7289f))
* file upload path encoding ([babd778](https://github.com/filebrowser/filebrowser/commit/babd7783afe85b790e1c558375d7b5013b2d366f))
* fix empty command name ([#1106](https://github.com/filebrowser/filebrowser/issues/1106)) ([36fb9f5](https://github.com/filebrowser/filebrowser/commit/36fb9f562a2c005ca4390fdebde0b4690201dff9))
* fix panic when accessing nonexistent .js file in static path ([#1105](https://github.com/filebrowser/filebrowser/issues/1105)) ([ad99bf1](https://github.com/filebrowser/filebrowser/commit/ad99bf180197e0e6d82231a86457585de16366a8))
* preview key shortcut conflict ([dd7b9dd](https://github.com/filebrowser/filebrowser/commit/dd7b9ddd8546361060ef99e838a691b2fc6c495a))
* search results absolute url ([26d62e4](https://github.com/filebrowser/filebrowser/commit/26d62e411716a5eb9a5a703e47484cfb3fbf3bd0))

## [2.7.0](https://github.com/filebrowser/filebrowser/compare/v2.6.2...v2.7.0) (2020-09-11)


### Features

* add --socket-perm flag to control unix socket file permissions (closes [#1060](https://github.com/filebrowser/filebrowser/issues/1060)) ([65ac734](https://github.com/filebrowser/filebrowser/commit/65ac73414fadc4686c94803a93ff319e8f7ce9d1))
* preview mobile dropdown ([7787344](https://github.com/filebrowser/filebrowser/commit/778734419de314d4cb64d07109bbab73f8e2e42a))
* preview size button ([3d2cb83](https://github.com/filebrowser/filebrowser/commit/3d2cb838d111ee61047599f49e76de80c821f341))
* put selected files in the root of the archive (closes [#1065](https://github.com/filebrowser/filebrowser/issues/1065)) ([8142b32](https://github.com/filebrowser/filebrowser/commit/8142b32f3865eccd3331328e0d087f805d186ed5))

### [2.6.2](https://github.com/filebrowser/filebrowser/compare/v2.6.1...v2.6.2) (2020-08-05)

### [2.6.1](https://github.com/filebrowser/filebrowser/compare/v2.6.0...v2.6.1) (2020-07-28)


### Bug Fixes

* delete cached previews when deleting file ([f5d02cd](https://github.com/filebrowser/filebrowser/commit/f5d02cdde97923b963878abf5a300393b9feb348))
* escape special characters in preview url (closes [#1002](https://github.com/filebrowser/filebrowser/issues/1002)) ([c9340af](https://github.com/filebrowser/filebrowser/commit/c9340af8d045671ad3338c5d2d887c335ab92de4))

## [2.6.0](https://github.com/filebrowser/filebrowser/compare/v2.5.0...v2.6.0) (2020-07-27)


### Features

* add lazy load of image thumbnails ([bc00165](https://github.com/filebrowser/filebrowser/commit/bc001650944ae963b12b5b2538a68de7cd0d8f82))
* add param to disable img resizing ([aa78e3a](https://github.com/filebrowser/filebrowser/commit/aa78e3ab1fcae6f618e811ba4e315a7a209f9df2))
* cache resized images ([95bc929](https://github.com/filebrowser/filebrowser/commit/95bc92955f391ece22c40d9592f2a3e6e26907b9))
* limit image resize workers ([94ef596](https://github.com/filebrowser/filebrowser/commit/94ef59602fb50fc21b1164feda90a3b9aeb5e972))


### Bug Fixes

* conflict handling on upload button ([f228fa5](https://github.com/filebrowser/filebrowser/commit/f228fa55408824618e9f0879da67c86d22b0d324))
* drop feedback ([f2d2c1c](https://github.com/filebrowser/filebrowser/commit/f2d2c1cbf85fba3edffb7b079f121ed3f0bc1e02))
* missing error message ([d9be370](https://github.com/filebrowser/filebrowser/commit/d9be370e2474b8070fa58db920c9481270cc4a48))
* parent verification on copy ([727c63b](https://github.com/filebrowser/filebrowser/commit/727c63b98e2964d0960d25914c296570f6c79478))
* path separator inconsistency on rename ([34dfb49](https://github.com/filebrowser/filebrowser/commit/34dfb49b719c948e709a4639b4af2c5cb73b3887))

## [2.5.0](https://github.com/filebrowser/filebrowser/compare/v2.4.0...v2.5.0) (2020-07-17)


### Features

* add previewer title and loading indicator ([716396a](https://github.com/filebrowser/filebrowser/commit/716396a726329f0ba42fc34167dd07497c5bf47c))
* duplicate files in the same directory ([43526d9](https://github.com/filebrowser/filebrowser/commit/43526d9d1a8c837245e3f5059e0b4737583eeaeb))
* file copy, move and paste conflict checking ([eed9da1](https://github.com/filebrowser/filebrowser/commit/eed9da1471723ed3fbe6c00b1d6362b1c5fd8b04))
* rename option on replace prompt ([2636f87](https://github.com/filebrowser/filebrowser/commit/2636f876ab8f88eea6d9548de524ca2339eb0843))
* upload queue ([6ec6a23](https://github.com/filebrowser/filebrowser/commit/6ec6a2386173410f5cab9941dbf1bacb6b70ddd2))


### Bug Fixes

* blinking previewer ([9a2ebba](https://github.com/filebrowser/filebrowser/commit/9a2ebbabe2e9f0c292701d33f36f9b7a457b1164))
* dark theme colors ([b3b6445](https://github.com/filebrowser/filebrowser/commit/b3b644527d5673e16e61d404ff58a3c7bd6b6637))
* directory conflict checking ([7e5beef](https://github.com/filebrowser/filebrowser/commit/7e5beeff464e75ab185c430cd96e7cc67209ccc1))
* prompt before closing window ([194030f](https://github.com/filebrowser/filebrowser/commit/194030fcfcf54a2cf5e2f8ececcbb4754474d8f8))
* remove incomplete uploaded files ([0727496](https://github.com/filebrowser/filebrowser/commit/0727496601a9918c8131c56f62419bfac7ac589a))
* reset clipboard after pasting cutted files ([10570ad](https://github.com/filebrowser/filebrowser/commit/10570ade442b573ebe00af08369e28b1b0688df6))

## [2.4.0](https://github.com/filebrowser/filebrowser/compare/v2.3.0...v2.4.0) (2020-07-07)


### Features

* full screen editor ([0d665e5](https://github.com/filebrowser/filebrowser/commit/0d665e528f880ceda0976ceed66070ac34de7969))


### Bug Fixes

* add preview bypass for .gif files ([#1012](https://github.com/filebrowser/filebrowser/issues/1012)) ([453636d](https://github.com/filebrowser/filebrowser/commit/453636dfe2bbf177c74617862eb763485d4774bf))
* prompt key shortcut conflict ([0d69fbd](https://github.com/filebrowser/filebrowser/commit/0d69fbd9a342aa2695859021df0c423e3ae4a4fa))

## [2.3.0](https://github.com/filebrowser/filebrowser/compare/v2.2.0...v2.3.0) (2020-06-26)


### Features

* add image thumbnails support ([#980](https://github.com/filebrowser/filebrowser/issues/980)) ([6b0d49b](https://github.com/filebrowser/filebrowser/commit/6b0d49b1fc8bdce89576ba91cc0b8ec594fcd625))


### Bug Fixes

* typo in image_templates (apline -> alpine) ([#1005](https://github.com/filebrowser/filebrowser/issues/1005)) ([84da110](https://github.com/filebrowser/filebrowser/commit/84da11008516a371fc0446d97863dc14d337aa25))

## [2.2.0](https://github.com/filebrowser/filebrowser/compare/v2.1.2...v2.2.0) (2020-06-22)


### Features

* add alpine and debian docker images ([66863b7](https://github.com/filebrowser/filebrowser/commit/66863b72f7664e6cb9417f7da542a92fa77ca635))
* add folder upload ([#981](https://github.com/filebrowser/filebrowser/issues/981)) ([8977344](https://github.com/filebrowser/filebrowser/commit/89773447a56675b298394149d7a05c5df4039f14)), closes [filebrowser/filebrowser#741](https://github.com/filebrowser/filebrowser/issues/741)
* add key shortcuts ([95316cb](https://github.com/filebrowser/filebrowser/commit/95316cbe8c8ac3dbb28310bc11ec347c0caf699b))
* upload progress based on total size ([#993](https://github.com/filebrowser/filebrowser/issues/993)) ([cd454ba](https://github.com/filebrowser/filebrowser/commit/cd454bae51f40b1249e6fa6133c2949970eb3018))


### Bug Fixes

* add a workaround to fix window freezing when viewing a large file [#992](https://github.com/filebrowser/filebrowser/issues/992) ([2412016](https://github.com/filebrowser/filebrowser/commit/241201657c2bf01806d02a297eb846b26102a479))
* apply all fs user rulles ([68f8348](https://github.com/filebrowser/filebrowser/commit/68f8348ddeecba570a361e7aba4546052cc3e356))
* frontend token validation ([dd40b0d](https://github.com/filebrowser/filebrowser/commit/dd40b0d9b9cc6268a611306ac4684a1af852b79d)), closes [filebrowser/filebrowser#638](https://github.com/filebrowser/filebrowser/issues/638)
* multiple selection count ([963837e](https://github.com/filebrowser/filebrowser/commit/963837ef1dc6e2e84fcf924606ce388ac30f3891))
* save event hook ([82c883f](https://github.com/filebrowser/filebrowser/commit/82c883f95eead9eebe215e230f74773c945f864a)), closes [filebrowser/filebrowser#696](https://github.com/filebrowser/filebrowser/issues/696)
