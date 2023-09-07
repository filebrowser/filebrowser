
interface file {
    name: string,
    modified: string,
    path: string,
    subtitles: any[],
    isDir: boolean,
    size: number,
    fullPath: string,
    type: uploadType
}

interface item {
    id: number,
    path: string,
    file: file,
    url?: string,
    dir?: boolean,
    from?: string,
    to?: string,
    name?: string,
    type?: uploadType
    overwrite: boolean
}

type uploadType = "video" | "audio" | "image" | "pdf" | "text" | "blob"

interface req {
    isDir?: boolean
}

interface uploads {
    [key: string]: upload
}

interface upload {
    id: number,
    file: file,
    type: string
}