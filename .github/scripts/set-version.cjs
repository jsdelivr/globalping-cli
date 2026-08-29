const { readFileSync, writeFileSync } = require('node:fs');

const versionFile = 'main.go';
const version = process.argv[2] ?? '';

if (!/^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?$/.test(version)) {
	throw new Error(`Invalid release version ${JSON.stringify(version)}.`);
}

const source = readFileSync(versionFile, 'utf8');
const versionPattern = /^(\s*version\s*=\s*)"[^"]+"$/m;

if (!versionPattern.test(source)) {
	throw new Error(`Unable to find the version variable in ${versionFile}.`);
}

writeFileSync(versionFile, source.replace(versionPattern, `$1"${version}"`));
