type apiUrl = string // Can also be set as a path eg: "path1" | "path2"

type resourcePath = string

type apiMethod =  "GET" | "POST" | "PUT" | "DELETE" | "PATCH"

// type apiContent = string | Blob | File 
type apiContent = Blob | File | Pick<ReadableStreamDefaultReader<any>, "read">  | ""

interface apiOpts {
    method?: apiMethod,
    headers?: object,
    body?: any
}

interface tusSettings {
    retryCount: number
}

type algo = any

type inline = any

interface share {
    expire: any,
    hash: string,
    path: string,
    userID: number,
    token: string
}

interface settings {
    any
}

type searchParams = any

