import ky from 'ky-universal';

import type {
    MeasurementResponse,
    PostMeasurement,
    PostMeasurementResponse,
} from './types';

export const getMeasurement = async (
    id: string
): Promise<MeasurementResponse> =>
    ky(`https://api.globalping.io/v1/measurements/${id}`).json();

export const postMeasurement = async (
    opts: PostMeasurement
): Promise<PostMeasurementResponse> =>
    ky
        .post('https://api.globalping.io/v1/measurements', {
            headers: {
                'Content-Type': 'application/json',
            },
            json: opts,
        })
        .json();
