"use strict";

const path = require("path");
const manifestFp = `${path.join(__dirname, "..", "public", "bundles")}${path.sep}manifest.json`;
const manifest = require(manifestFp);

for (const key of Object.keys(manifest)) manifest[key] = manifest[key].split("/").pop();

require("fs").writeFileSync(manifestFp, JSON.stringify(manifest, undefined, "\t"));
