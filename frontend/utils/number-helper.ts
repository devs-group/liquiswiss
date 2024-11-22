export const AmountToInteger = (amount: number) => {
    return Math.round(amount * 100)
}

export const AmountToFloat = (amount: number) => {
    return Math.round(amount / 100 * 100) / 100
}

export const parseCurrency = (input: string|number) => {
    if (typeof input === 'number') {
        input = input.toString(10)
    }

    // Replace all commas with dots (unifying decimal separator)
    let unifiedInput = input.replace(/,/g, '.');

    // Remove all invalid characters except numbers and dots
    unifiedInput = unifiedInput.replace(/[^0-9.]/g, '');

    // Ensure only one decimal separator is allowed
    let parts = unifiedInput.split('.');
    if (parts.length > 2) {
        // Keep the last detected decimal part
        const decimals = parts.pop();
        // Reassemble with a single dot
        unifiedInput = parts.join('') + '.' + decimals;
    }

    return unifiedInput;
}
