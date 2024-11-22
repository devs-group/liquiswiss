export const scrollToParentBottom = (elementID: string) => {
    const parentElement = document.getElementById(elementID)?.parentElement;
    if (parentElement) {
        parentElement.scrollTo({
            top: parentElement.scrollHeight,
            behavior: 'smooth'
        });
    }
}

export const parseNumberInput = (event: InputEvent, amount: Ref<number>) => {
    const element = event.target as HTMLInputElement
    const cursorPosition = element.selectionStart || 0;
    const lengthBefore = element.value.length;

    const parsedAmount = parseCurrency(amount.value)
    if (!parsedAmount.endsWith('.')) {
        amount.value = parseFloat(parsedAmount)

        const lengthAfter = amount.value.toString().length;
        const offset = lengthAfter - lengthBefore;
        const newCursorPosition = cursorPosition + offset;

        nextTick(() => element.setSelectionRange(newCursorPosition, newCursorPosition))
    }
}