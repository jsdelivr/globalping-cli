import { cac } from 'cac';

import { version } from '../package.json';
import { render } from './render';

const cli = cac('globalping');

cli.command('ping <target> from [...locations]', 'Use ping command')
    .option('--from <location>', 'Probe locations')
    .option('--limit <count>', 'Number of probes')
    .option('--packets <count>', 'Number of packets')
    .option('--ci', 'Disable pretty rendering')
    .action(async (target, locationArr, opts) => {
        try {
            const args = {
                cmd: 'ping',
                target,
                locationArr,
                ...opts,
            };

            console.log(JSON.stringify(args));
            render(args);
        } catch (error) {
            console.error(error);
        }
    });

cli.command('traceroute <target> from [...locations]', 'Use traceroute command')
    .option('--from <location>', 'Probe locations')
    .option('--limit <count>', 'Number of probes')
    .option('--protocol <type>', 'Protocol to use')
    .option('--port <number>', 'Port to use')
    .action(async (target, locationArr, opts) => {
        try {
            const args = {
                cmd: 'traceroute',
                target,
                locationArr,
                ...opts,
            };

            render(args);
        } catch (error) {
            console.error(error);
        }
    });

cli.command('dns <target> from [...locations]', 'Use DNS command')
    .option('--from <location>', 'Probe locations')
    .option('--limit <count>', 'Number of probes')
    .option('--query <type>', 'Query type')
    .option('--port <number>', 'Port to use')
    .option('--protocol <type>', 'Protocol to use')
    .option('--resolver <address>', 'Use resolver')
    .option('--trace', 'Use trace')
    .action(async (target, locationArr, opts) => {
        try {
            const args = {
                cmd: 'dns',
                target,
                locationArr,
                ...opts,
            };

            render(args);
        } catch (error) {
            console.error(error);
        }
    });

cli.command('mtr <target> from [...locations]', 'Use MTR command')
    .option('--from <location>', 'Probe locations')
    .option('--protocol <type>', 'Protocol to use')
    .option('--port <number>', 'Use port number')
    .option('--packets <count>', 'Number of packets')
    .action(async (target, locationArr, opts) => {
        try {
            const args = {
                cmd: 'mtr',
                target,
                locationArr,
                ...opts,
            };

            render(args);
        } catch (error) {
            console.error(error);
        }
    });

cli.command('http <target> from [...locations]')
    .option('--from <location>', 'Probe locations')
    .option('--port <number>', 'Port number')
    .option('--protocol <type>', 'Protocol to use')
    .option('--path <route>', 'Path to use')
    .option('--query <string>', 'Query to use')
    .option('--method <type>', 'HTTP method to use')
    .option('--host <string>', 'Hostname to use')
    .option('--headers <string>', 'Headers to use')
    .action(async (target, locationArr, opts) => {
        try {
            const args = {
                cmd: 'http',
                target,
                locationArr,
                ...opts,
            };

            render(args);
        } catch (error) {
            console.error(error);
        }
    });

cli.help();
cli.version(version);

cli.parse();
