"use strict";

const path = require("path");
const cwd = path.join(__dirname, "..");
const child_process = require("child_process");

const platforms = [
    {
        os: "darwin",
        arch: "arm64",
        fileSuffix: "",
    },
    {
        os: "darwin",
        arch: "amd64",
        fileSuffix: "",
    },
    {
        os: "freebsd",
        arch: "386",
        fileSuffix: "",
    },
    {
        os: "freebsd",
        arch: "amd64",
        fileSuffix: "",
    },
    {
        os: "freebsd",
        arch: "arm",
        fileSuffix: "",
    },
    {
        os: "linux",
        arch: "386",
        fileSuffix: "",
    },
    {
        os: "linux",
        arch: "amd64",
        fileSuffix: "",
    },
    {
        os: "linux",
        arch: "arm",
        fileSuffix: "",
    },
    {
        os: "windows",
        arch: "386",
        fileSuffix: ".exe",
    },
    {
        os: "windows",
        arch: "amd64",
        fileSuffix: ".exe",
    },
    {
        os: "windows",
        arch: "arm",
        fileSuffix: ".exe",
    },
];

const res = child_process.spawnSync("go", ["mod", "download"],
    {stdio: [process.stdin, process.stdout, process.stderr], cwd});
if (res.status !== 0) throw new Error(`Got status code ${res.status}`);

const runPromise = (cmd, args, env) => new Promise((res, rej) => {
    let procEnv = process.env;
    if (env) procEnv = Object.assign({}, procEnv, env);
    const proc = child_process.spawn(cmd, args,
        {stdio: [process.stdin, process.stdout, process.stderr], env: procEnv, cwd});
    proc.on("error", err => rej(err));
    proc.on("exit", code => {
        if (code !== 0) throw new Error(`Process exited with code ${code}.`);
        res();
    });
});

(async () => {
    // Build the spherical binaries.
    try {
        await Promise.all(platforms.map(platform => runPromise("go", [
            "build", "-o", `./dist/spherical-${platform.os}-${platform.arch}${platform.fileSuffix}`,
            "./cmd/spherical",
        ], {GOOS: platform.os, GOARCH: platform.arch})));
    } catch (e) {
        console.error("Failed to build binaries", e);
        process.exit(1);
    }
    console.log("Spherical binaries built!");

    // Build the hash validator binaries.
    try {
        await Promise.all(platforms.map(platform => runPromise("go", [
            "build", "-o", `./dist/hashvalidate-${platform.os}-${platform.arch}${platform.fileSuffix}`,
            "./cmd/hashvalidate",
        ], {GOOS: platform.os, GOARCH: platform.arch})));
    } catch (e) {
        console.error("Failed to build binaries", e);
        process.exit(1);
    }
    console.log("Hash validator binaries built!");
})();
