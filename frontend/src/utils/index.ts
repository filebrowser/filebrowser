import { partial } from "filesize";

/**
 * Formats filesize as KiB/MiB/...
 */
export const filesize = partial({ base: 2 });
