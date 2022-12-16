export const throwArgError = (
    invalid: string | undefined,
    type: string,
    expected: string
) => {
    throw new TypeError(
        `Invalid argument "${invalid}" for "${type}"!\nExpected "${expected}".`
    );
};
