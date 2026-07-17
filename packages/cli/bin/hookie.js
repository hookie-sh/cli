#!/usr/bin/env node
/**
 * Hookie CLI wrapper: spawns the binary downloaded by postinstall.
 * Requires Node 18+.
 */

const { spawnSync } = require("child_process");
const path = require("path");

const binaryName = process.platform === "win32" ? "hookie.exe" : "hookie";
const binaryPath = path.join(__dirname, binaryName);

const result = spawnSync(binaryPath, process.argv.slice(2), {
  stdio: "inherit",
  windowsHide: true,
});

process.exit(result.status ?? (result.signal ? 128 + result.signal : 0));
