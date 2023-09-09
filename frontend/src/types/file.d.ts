interface IFile {
    index?: number
    name: string,
    modified: string,
    path: string,
    subtitles: any[],
    isDir: boolean,
    size: number,
    fullPath: string,
    type: uploadType,
    items: IFile[]
    token?: string,
    hash: string,
    url?: string
}



type uploadType = "video" | "audio" | "image" | "pdf" | "text" | "blob" | "textImmutable"

type req = {
    path: string
    name: string
    size: number
    extension: string
    modified: string
    mode: number
    isDir: boolean
    isSymlink: boolean
    type: string
    url: string
    hash: string
  }
  


interface uploads {
    [key: string]: upload
}

interface upload {
    id: number,
    file: file,
    type: string
}