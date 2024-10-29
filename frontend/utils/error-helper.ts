import {FetchError} from "ofetch";

export const IsAbortedError = (error: FetchError|null) => {
    return error?.cause && (error.cause as { name: string }).name === "AbortError"
}