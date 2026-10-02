// Learn more https://docs.expo.dev/guides/monorepos/
const path = require("node:path");
const { getDefaultConfig } = require("expo/metro-config");

const config = getDefaultConfig(__dirname);

// Watch the repo-level shared/ packages, which live outside mobile/.
config.watchFolders = [...(config.watchFolders ?? []), path.resolve(__dirname, "../../../shared")];

module.exports = config;
