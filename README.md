# invert-image

Invert the colors of a PNG or JPEG image.

```bash
invert-image photo.png inverted.png
invert-image photo.jpg
```

Without an output filename, the original is replaced. An explicit output filename
must end in `.png`, `.jpg`, or `.jpeg` and selects the output format.

PNG transparency and 16-bit color depth are preserved. JPEG is re-encoded at
quality 95; transparent pixels are flattened onto white when converting to JPEG.
Image metadata is not retained.

Use `--help` for usage or `--version` for the build version.

## Download

Download a Linux or Windows build from
[GitHub Releases](https://github.com/dannyben/invert-image/releases).

## Development

```bash
op run photo.png inverted.png
op check
op build
op build-windows
```

Builds are written to `dist/`. Versions come from `vX.Y.Z` Git tags, with development
and dirty suffixes between releases. `op release-check` validates GoReleaser;
`op release-snapshot` builds local archives. Pushing a version tag runs the release
workflow for Linux and Windows (amd64 and arm64).
