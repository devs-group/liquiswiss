import {da} from "cronstrue/dist/i18n/locales/da";

export const DateStringToFormattedDate = (date: string|Date) => {
    const fmt = Intl.DateTimeFormat('de-CH', {day: '2-digit', month: '2-digit', year: 'numeric', timeZone: 'UTC'})
    return fmt.format(date instanceof Date ? date : new Date(date))
}

export const DateToUTCDate = (date: string|Date) => {
    let dateToFormat = date instanceof Date ? date : new Date(date);
    return new Date(dateToFormat.toLocaleDateString('en-US', {timeZone: 'UTC'}))
}

export const DateToApiFormat = (date: string | Date) => {
    const dateToFormat = date instanceof Date ? date : new Date(date);

    const year = dateToFormat.getUTCFullYear(); // Use UTC methods to avoid time zone shift
    const month = (dateToFormat.getUTCMonth() + 1).toString().padStart(2, '0'); // Ensure two digits for the month
    const day = dateToFormat.getUTCDate().toString().padStart(2, '0'); // Ensure two digits for the day

    return `${year}-${month}-${day}`;
};

export const NumberToFormattedCurrency = (amount: number, locale: string) => {
    const fmt = Intl.NumberFormat(locale)
    return fmt.format(amount)
}