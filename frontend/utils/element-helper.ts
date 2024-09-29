export const scrollToParentBottom = (elementID: string) => {
    const parentElement = document.getElementById(elementID)?.parentElement;
    if (parentElement) {
        parentElement.scrollTo({
            top: parentElement.scrollHeight,
            behavior: 'smooth'
        });
    }
}