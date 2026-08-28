const releaseNoteTypes = [
	{ type: 'feat', section: 'Features' },
	{ type: 'fix', section: 'Bug Fixes' },
	{ type: 'perf', section: 'Performance Improvements' },
	{ type: 'revert', section: 'Reverts' },
	{ type: 'docs', section: 'Documentation', hidden: true },
	{ type: 'style', section: 'Styles', hidden: true },
	{ type: 'chore', section: 'Miscellaneous Chores', hidden: true },
	{ type: 'refactor', section: 'Code Refactoring', hidden: true },
	{ type: 'test', section: 'Tests', hidden: true },
	{ type: 'build', section: 'Build System', hidden: true },
	{ type: 'ci', section: 'Continuous Integration', hidden: true },
	{ type: 'misc', section: 'Miscellaneous', hidden: true },
];

module.exports = {
	branches: [ 'master' ],
	tagFormat: 'v${version}',
	repositoryUrl: 'git@github.com:jsdelivr/globalping-cli.git',
	plugins: [
		[ '@semantic-release/commit-analyzer', {
			preset: 'conventionalcommits',
			releaseRules: [
				{ type: 'misc', release: 'patch' },
			],
		}],
		[ '@semantic-release/release-notes-generator', {
			preset: 'conventionalcommits',
			presetConfig: {
				types: releaseNoteTypes,
			},
		}],
		[ '@semantic-release/exec', {
			prepareCmd: 'node .github/scripts/set-version.cjs ${nextRelease.version}',
		}],
		'@semantic-release/github',
		[ '@semantic-release/git', {
			assets: [ 'main.go' ],
			message: 'chore: [skip ci] bump version to ${nextRelease.version}',
		}],
	],
};
