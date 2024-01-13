#!/bin/bash

svg_input="$1"
destination_folder="${2:-.}"

if [ -z "$svg_input" ]; then
  echo "Usage: $0 <input_svg_file> [destination_folder]"
  echo "Converts an SVG input to various image formats and sizes based on a predefined list."
  exit 1
fi

if [ ! -d "$destination_folder" ]; then
  echo "Destination folder does not exist. Creating: $destination_folder"
  mkdir -p "$destination_folder"
fi

entries=(
  "android-chrome-192x192.png"
  "android-chrome-512x512.png"
  "apple-touch-icon.png"
  "favicon-16x16.png"
  "favicon-32x32.png"
  "favicon.ico"
  "mstile-144x144.png"
  "mstile-150x150.png"
  "mstile-310x150.png"
  "mstile-310x310.png"
  "mstile-70x70.png"
  "safari-pinned-tab.svg"
)

for entry in "${entries[@]}"; do
  name="${entry%.*}"
  extension="${entry##*.}"
  output="$destination_folder/${name}_output.${extension}"

  # Convert SVG to specified dimensions and format
  case "$extension" in
    "png")
      convert -background none -resize "${name: -4}"x"${name: -4}" "$svg_input" "$output"
      ;;
    "ico")
      convert "$svg_input" -bordercolor white -border 0 -alpha off "$output"
      ;;
    "svg")
      # SVGs don't need conversion, just copy
      cp "$svg_input" "$output"
      ;;
    *)
      echo "Unsupported file format: $entry"
      continue
      ;;
  esac

  echo "Converted $svg_input to $output"
done
