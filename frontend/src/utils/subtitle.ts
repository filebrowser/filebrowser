import { StatusError } from "@/api/utils";

export async function loadSubtitle(subUrl: string) {
  let url: URL;
  try {
    url = new URL(subUrl);
  } catch (_) {
    // treat it as a relative url
    // we only need this for filename
    url = new URL(subUrl, window.location.origin);
  }

  const label = decodeURIComponent(
    url.pathname
      .split("/")
      .pop()!
      .replace(/\.[^/.]+$/, "")
  );
  let src;
  if (url.pathname.toLowerCase().endsWith(".srt")) {
    try {
      const resp = await fetch(subUrl);
      if (!resp.ok) {
        throw new StatusError(
          `Failed to fetch subtitle from ${subUrl}!`,
          resp.status
        );
      }
      const vtt = srtToVttBlob(await resp.text());
      src = URL.createObjectURL(vtt);
    } catch (error) {
      console.error(error);
    }
  } else if (url.pathname.toLowerCase().endsWith(".vtt")) {
    src = subUrl;
  }
  return { src, label };
}

export function srtToVttBlob(srtData: string) {
  const VTT_HEAD = "WEBVTT\n\n";
  // Replace line breaks with \n
  let subtitles = srtData.replace(/\r\n|\r|\n/g, "\n");
  // commas -> dots in timestamps
  subtitles = subtitles.replace(/(\d\d:\d\d:\d\d),(\d\d\d)/g, "$1.$2");
  // map SRT font colors to VTT cue span classes
  const colorMap: Record<string, string> = {};
  // font tags -> ::cue span tags
  subtitles = subtitles.replace(
    /<font color="([^"]+)">([\s\S]*?)<\/font>/g,
    function (_match, color, text) {
      const key =
        "c_" + color.replace(/^rgb/, "").replace(/\W/g, "").toLowerCase();
      colorMap[key] = color;
      return `<c.${key}>${text.replace("\n", "").trim()}</c>`;
    }
  );
  subtitles = subtitles.replace(/<br\s*\/?>/g, "\n");
  let vttSubtitles = VTT_HEAD;
  if (Object.keys(colorMap).length) {
    let vttStyles = "";
    for (const cssClass in colorMap) {
      const color = colorMap[cssClass];
      // add cue style declaration
      vttStyles += `::cue(.${cssClass}) {color: ${color};}\n`;
    }
    vttSubtitles += vttStyles;
  }
  vttSubtitles += "\n"; // an empty line MUST separate styles from subs
  vttSubtitles += subtitles;
  return new Blob([vttSubtitles], { type: "text/vtt" });
}
