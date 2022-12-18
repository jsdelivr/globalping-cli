/* eslint-disable no-await-in-loop */
import { render as ink } from 'ink';

import { getMeasurement, postMeasurement } from './api';
import { parseArgs } from './parse';
import type { Arguments } from './types';

const noink = async (args: Arguments) => {
    console.log(parseArgs(args));
    const { id } = await postMeasurement(parseArgs(args));
    let res = await getMeasurement(id);
    // CI doesn't need real time results, just a final answer
    while (res.status === 'in-progress') {
        // eslint-disable-next-line no-promise-executor-return
        await new Promise((resolve) => setTimeout(resolve, 100));
        res = await getMeasurement(id);
    }

    if (args.json) {
        console.log(JSON.stringify(res));
    } else {
        const { results } = res;
        for (const result of results) {
            const msg = `${result.probe.continent}, ${result.probe.country}, ${
                result.probe.state ? `(${result.probe.state}), ` : ''
            }${result.probe.city}, ASN:${result.probe.asn}`;
            console.log(msg);
            console.log(result.result.rawOutput);
        }
    }
};

export const render = async (args: Arguments) => {
    if ('ci' in args || 'json' in args) return noink(args);

    noink(args);
};
