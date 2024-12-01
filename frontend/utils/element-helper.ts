import {parseCurrency} from "~/utils/number-helper";

export const scrollToParentBottom = (elementID: string) => {
    const parentElement = document.getElementById(elementID)?.parentElement;
    if (parentElement) {
        parentElement.scrollTo({
            top: parentElement.scrollHeight,
            behavior: 'smooth'
        });
    }
}

export const parseNumberInput = (event: InputEvent, amount: Ref<number>, allowNegative: boolean) => {
    const element = event.target as HTMLInputElement
    const cursorPosition = element.selectionStart || 0;
    const lengthBefore = element.value.length;

    const parsedAmount = parseCurrency(amount.value, allowNegative)
    if (!parsedAmount.endsWith('.')) {
        nextTick(() => {
            amount.value = parsedAmount.length > 0 ? parseFloat(parsedAmount) : 0
        }).then(() => {
            const lengthAfter = amount.value.toString().length;
            const offset = lengthAfter - lengthBefore;
            const newCursorPosition = cursorPosition + offset;

            element.setSelectionRange(newCursorPosition, newCursorPosition)
        })
    }
}