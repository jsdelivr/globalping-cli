import {
    ALLOWED_DNS_PROTOCOLS,
    ALLOWED_DNS_TYPES,
    ALLOWED_HTTP_METHODS,
    ALLOWED_HTTP_PROTOCOLS,
    ALLOWED_MTR_PROTOCOLS,
    ALLOWED_QUERY_TYPES,
    ALLOWED_TRACE_PROTOCOLS,
    Arguments,
    isDnsProtocol,
    isDnsType,
    isHttpMethod,
    isHttpProtocol,
    isMtrProtocol,
    isTraceProtocol,
    PostMeasurement,
} from './types';
import { throwArgError } from './utils';

export const parseArgs = (args: Arguments): PostMeasurement => {
    const {
        cmd,
        target,
        from,
        locationArr,
        limit,
        packets,
        protocol,
        port,
        query,
        resolver,
        trace,
        path,
        method,
        host,
        headers,
    } = args;

    // When using the ping x from y format, the from is picked up in the locations arr
    if (locationArr[0] === 'from') locationArr.shift();

    const locations = [{ magic: locationArr.join(' ') ?? from }];

    if (cmd === 'ping')
        return {
            type: 'ping',
            target,
            limit: limit ?? 1,
            locations,
            measurementOptions: {
                ...(packets && { packets }),
            },
        };

    if (cmd === 'traceroute')
        return {
            type: 'traceroute',
            target,
            limit: limit ?? 1,
            locations,
            measurementOptions: {
                ...(protocol && {
                    protocol: isTraceProtocol(protocol)
                        ? protocol
                        : throwArgError(
                              protocol,
                              'protocol',
                              [...ALLOWED_TRACE_PROTOCOLS].join(', ')
                          ),
                }),
                ...(port && { port }),
            },
        };

    if (cmd === 'dns')
        return {
            type: 'dns',
            target,
            limit: limit ?? 1,
            locations,
            measurementOptions: {
                ...(query && {
                    query: {
                        type: isDnsType(query)
                            ? query
                            : throwArgError(
                                  query,
                                  'query',
                                  [...ALLOWED_DNS_TYPES].join(', ')
                              ),
                    },
                }),
                ...(protocol && {
                    protocol: isDnsProtocol(protocol)
                        ? protocol
                        : throwArgError(
                              protocol,
                              'protocol',
                              [...ALLOWED_DNS_PROTOCOLS].join(', ')
                          ),
                }),
                ...(port && { port }),
                ...(resolver && { resolver }),
                ...(trace && { trace }),
            },
        };

    if (cmd === 'mtr') {
        return {
            type: 'mtr',
            target,
            limit: limit ?? 1,
            locations,
            measurementOptions: {
                ...(protocol && {
                    protocol: isMtrProtocol(protocol)
                        ? protocol
                        : throwArgError(
                              protocol,
                              'protocol',
                              [...ALLOWED_MTR_PROTOCOLS].join(', ')
                          ),
                }),
                ...(port && { port }),
                ...(packets && { packets }),
            },
        };
    }

    if (cmd === 'http')
        return {
            type: 'http',
            target,
            limit: limit ?? 1,
            locations,
            measurementOptions: {
                ...(port && { port }),
                ...(protocol && {
                    protocol: isHttpProtocol(protocol)
                        ? protocol
                        : throwArgError(
                              protocol,
                              'protocol',
                              [...ALLOWED_HTTP_PROTOCOLS].join(', ')
                          ),
                }),
                request: {
                    ...(path && { path }),
                    ...(query && { query }),
                    ...(method && {
                        method: isHttpMethod(method)
                            ? method
                            : throwArgError(
                                  method,
                                  'method',
                                  [...ALLOWED_HTTP_METHODS].join(', ')
                              ),
                    }),
                    ...(host && { host }),
                    ...(headers && { headers }),
                },
            },
        };

    throwArgError(String(cmd), 'command', [...ALLOWED_QUERY_TYPES].join(', '));
    throw new Error('Unknown error.');
};
