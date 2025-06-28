#!/bin/sh

# Initial setup
cp /tmp/README.md /build/docs/index.md
sed -i 's|site/docs/||g' /build/docs/index.md

# Start mkdocs in the background
mkdocs serve -a 0.0.0.0:8000 --dirtyreload &

# Watch for changes in README.md
while true; do
  inotifywait -e modify /tmp/README.md
  echo "README.md changed. Updating index.md..."
  cp /tmp/README.md /build/docs/index.md
  sed -i 's|site/docs/||g' /build/docs/index.md
done

