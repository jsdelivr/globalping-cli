/* eslint-disable no-await-in-loop */
import { cac } from 'cac';

import { version } from '../package.json';
import { getMeasurement, postMeasurement } from './api';
import { parseArgs } from './parse';

const cli = cac('globalping');

cli.command('ping <target> from [...locations]', 'Use ping command')
    .option('--from <location>', 'Probe locations')
    .option('--limit <count>', 'Number of probes')
    .option('--packets <count>', 'Number of packets')
    .option('--ci', 'Disable pretty rendering')
    .action(async (target, locations, opts) => {
        try {
            const args = {
                cmd: 'ping',
                target,
                locations,
                ...opts,
            };

            console.log(JSON.stringify(args));

            if ('ci' in opts) {
                const { id } = await postMeasurement(parseArgs(args));
                let res = await getMeasurement(id);
                // CI doesn't need real time results, just a final answer
                while (res.status === 'in-progress') {
                    // eslint-disable-next-line no-promise-executor-return
                    await new Promise((resolve) => setTimeout(resolve, 100));
                    res = await getMeasurement(id);
                }
            } else {
                // render
            }
        } catch (error) {
            console.log(error);
        }
    });

cli.command('traceroute [target]', 'Use traceroute command')
    .option('-F --from', 'Probe locations')
    .option('-L --limit', 'Number of probes')
    .option('--protocol', 'Protocol to use')
    .option('--port', 'Port to use');

cli.command('dns [target]', 'Use DNS command')
    .option('-F --from', 'Probe locations')
    .option('-L --limit', 'Number of probes')
    .option('--query', 'Query type')
    .option('--port', 'Port to use')
    .option('--protocol', 'Protocol to use');

cli.help();
cli.version(version);

cli.parse();
